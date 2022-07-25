package lefthook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/runner"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	envEnabled    = "LEFTHOOK"       // "0", "false"
	envSkipOutput = "LEFTHOOK_QUIET" // "pre-commit,post-commit"

	skipMeta    = 0b0001
	skipSuccess = 0b0010
	skipFailure = 0b0100
	skipSummary = 0b1000
)

type skipOutputSettings int8

func (s skipOutputSettings) doSkip(option int8) bool {
	return int8(s)&option != 0
}

func Run(opts *Options, hookName string, gitArgs []string) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Run(hookName, gitArgs)
}

func (l *Lefthook) Run(hookName string, gitArgs []string) error {
	if os.Getenv(envEnabled) == "0" || os.Getenv(envEnabled) == "false" {
		return nil
	}

	if hookName == config.GhostHookName {
		log.SetLevel(log.WarnLevel)
	}

	// Load config
	cfg, err := config.Load(l.Fs, l.repo.RootPath)
	if err != nil {
		return err
	}
	if err = cfg.Validate(); err != nil {
		return err
	}

	if tags := os.Getenv(envSkipOutput); tags != "" {
		cfg.SkipOutput = append(cfg.SkipOutput, strings.Split(tags, ",")...)
	}

	var outputSettings skipOutputSettings
	for _, param := range cfg.SkipOutput {
		switch param {
		case "meta":
			outputSettings |= skipMeta
		case "success":
			outputSettings |= skipSuccess
		case "failure":
			outputSettings |= skipFailure
		case "summary":
			outputSettings |= skipSummary
		}
	}

	if cfg.Colors != config.DefaultColorsEnabled {
		log.SetColors(cfg.Colors)
	}

	if !outputSettings.doSkip(skipMeta) {
		log.Info(log.Cyan("Lefthook v" + version.Version(false)))
	}

	// This line controls updating the git hook if config has changed
	if err = l.createHooksIfNeeded(cfg, false); err != nil {
		log.Warn(
			`⚠️ There was a problem with synchronizing git hooks.
Run 'lefthook install' manually.`,
		)
	}

	if !outputSettings.doSkip(skipMeta) {
		log.Info(log.Cyan("RUNNING HOOK:"), log.Bold(hookName))
	}

	// Find the hook
	hook, ok := cfg.Hooks[hookName]
	if !ok {
		return nil
	}
	if err := hook.Validate(); err != nil {
		return err
	}

	startTime := time.Now()
	resultChan := make(chan runner.Result, len(hook.Commands)+len(hook.Scripts))
	run := runner.NewRunner(l.Fs, l.repo, hook, gitArgs, resultChan)

	go func() {
		run.RunAll(
			[]string{
				filepath.Join(cfg.SourceDir, hookName),
				filepath.Join(cfg.SourceDirLocal, hookName),
			},
		)
		close(resultChan)
	}()

	var results []runner.Result
	for res := range resultChan {
		results = append(results, res)
	}

	if !outputSettings.doSkip(skipSummary) {
		printSummary(time.Since(startTime), results, outputSettings)
	}

	for _, result := range results {
		if result.Status == runner.StatusErr {
			return errors.New("") // No error should be printed
		}
	}

	return nil
}

func printSummary(
	duration time.Duration,
	results []runner.Result,
	outputSettings skipOutputSettings,
) {
	if len(results) == 0 {
		log.Info(log.Cyan("\nSUMMARY: (SKIP EMPTY)"))
		return
	}

	log.Info(log.Cyan(
		fmt.Sprintf("\nSUMMARY: (done in %.2f seconds)", duration.Seconds()),
	))

	if !outputSettings.doSkip(skipSuccess) {
		for _, result := range results {
			if result.Status != runner.StatusOk {
				continue
			}

			log.Infof("✔️  %s\n", log.Green(result.Name))
		}
	}

	if !outputSettings.doSkip(skipFailure) {
		for _, result := range results {
			if result.Status != runner.StatusErr {
				continue
			}

			var failText string
			if len(result.Text) != 0 {
				failText = fmt.Sprintf(": %s", result.Text)
			}

			log.Infof("🥊  %s%s\n", log.Red(result.Name), log.Red(failText))
		}
	}
}
