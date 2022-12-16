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
	envEnabled    = "LEFTHOOK"         // "0", "false"
	envSkipOutput = "LEFTHOOK_QUIET"   // "meta,success,failure,summary,execution"
	envVerbose    = "LEFTHOOK_VERBOSE" // keep all output
)

type RunArgs struct {
	NoTTY bool
}

func Run(opts *Options, args RunArgs, hookName string, gitArgs []string) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Run(hookName, args, gitArgs)
}

func (l *Lefthook) Run(hookName string, args RunArgs, gitArgs []string) error {
	if os.Getenv(envEnabled) == "0" || os.Getenv(envEnabled) == "false" {
		return nil
	}

	var verbose bool
	if l.Verbose || os.Getenv(envVerbose) == "1" || os.Getenv(envVerbose) == "true" {
		log.SetLevel(log.DebugLevel)
		verbose = true
	}

	// Load config
	cfg, err := config.Load(l.Fs, l.repo)
	if err != nil {
		return err
	}
	if err = cfg.Validate(); err != nil {
		return err
	}

	// Suppress prepare-commit-msg output if the hook doesn't exists in config.
	// prepare-commit-msg hook is used for seemless synchronization of hooks with config.
	// See: internal/lefthook/install.go
	_, ok := cfg.Hooks[hookName]
	if hookName == config.GhostHookName && !ok && !verbose {
		log.SetLevel(log.WarnLevel)
	}

	if tags := os.Getenv(envSkipOutput); tags != "" {
		cfg.SkipOutput = append(cfg.SkipOutput, strings.Split(tags, ",")...)
	}

	var logSettings log.SkipSettings
	for _, skipOption := range cfg.SkipOutput {
		(&logSettings).ApplySetting(skipOption)
	}

	if cfg.Colors != config.DefaultColorsEnabled {
		log.SetColors(cfg.Colors)
	}

	if !logSettings.SkipMeta() {
		log.Info(log.Cyan("Lefthook v" + version.Version(false)))
	}

	// This line controls updating the git hook if config has changed
	if err = l.createHooksIfNeeded(cfg, false); err != nil {
		log.Warn(
			`⚠️  There was a problem with synchronizing git hooks.
Run 'lefthook install' manually.`,
		)
	}

	if !logSettings.SkipMeta() {
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

	run := runner.NewRunner(
		runner.Opts{
			Fs:           l.Fs,
			Repo:         l.repo,
			Hook:         hook,
			HookName:     hookName,
			GitArgs:      gitArgs,
			ResultChan:   resultChan,
			SkipSettings: logSettings,
			DisableTTY:   cfg.NoTTY || args.NoTTY,
		},
	)

	sourceDirs := []string{
		filepath.Join(l.repo.RootPath, cfg.SourceDir),
		filepath.Join(l.repo.RootPath, cfg.SourceDirLocal),
	}

	if cfg.Remote.Configured() {
		// Apend only source_dir, because source_dir_local doesn't make sense
		sourceDirs = append(
			sourceDirs,
			filepath.Join(
				l.repo.RemoteFolder(cfg.Remote.GitURL),
				cfg.SourceDir,
			),
		)
	}

	go func() {
		run.RunAll(sourceDirs)
		close(resultChan)
	}()

	var results []runner.Result
	for res := range resultChan {
		results = append(results, res)
	}

	if !logSettings.SkipSummary() {
		printSummary(time.Since(startTime), results, logSettings)
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
	logSettings log.SkipSettings,
) {
	if len(results) == 0 {
		log.Info(log.Cyan("\nSUMMARY: (SKIP EMPTY)"))
		return
	}

	log.Info(log.Cyan(
		fmt.Sprintf("\nSUMMARY: (done in %.2f seconds)", duration.Seconds()),
	))

	if !logSettings.SkipSuccess() {
		for _, result := range results {
			if result.Status != runner.StatusOk {
				continue
			}

			log.Infof("✔️  %s\n", log.Green(result.Name))
		}
	}

	if !logSettings.SkipFailure() {
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
