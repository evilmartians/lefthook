package lefthook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/run"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	envEnabled    = "LEFTHOOK"       // "0", "false"
	envSkipOutput = "LEFTHOOK_QUIET" // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
)

type RunArgs struct {
	NoTTY           bool
	AllFiles        bool
	Files           []string
	RunOnlyCommands []string
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
	if l.Verbose {
		log.SetLevel(log.DebugLevel)
		verbose = true
	}

	// Load config
	cfg, err := config.Load(l.Fs, l.repo)
	if err != nil {
		var notFoundErr config.NotFoundError
		if ok := errors.As(err, &notFoundErr); ok {
			log.Warn(err.Error())
			return nil
		}

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

	if !logSettings.SkipMeta() {
		log.Box(
			log.Cyan("🥊 lefthook ")+log.Gray(fmt.Sprintf("v%s", version.Version(false))),
			log.Gray("hook: ")+log.Bold(hookName),
		)
	}

	// This line controls updating the git hook if config has changed
	if err = l.createHooksIfNeeded(cfg, false); err != nil {
		log.Warn(
			`⚠️  There was a problem with synchronizing git hooks.
Run 'lefthook install' manually.`,
		)
	}

	// Find the hook
	hook, ok := cfg.Hooks[hookName]
	if !ok {
		if slices.Contains(config.AvailableHooks[:], hookName) {
			log.Debugf("[lefthook] skip: Hook %s doesn't exist in the config", hookName)
			return nil
		}

		return fmt.Errorf("Hook %s doesn't exist in the config", hookName)
	}
	if err := hook.Validate(); err != nil {
		return err
	}

	startTime := time.Now()
	resultChan := make(chan run.Result, len(hook.Commands)+len(hook.Scripts))

	runner := run.NewRunner(
		run.Options{
			Repo:            l.repo,
			Hook:            hook,
			HookName:        hookName,
			GitArgs:         gitArgs,
			ResultChan:      resultChan,
			SkipSettings:    logSettings,
			DisableTTY:      cfg.NoTTY || args.NoTTY,
			AllFiles:        args.AllFiles,
			Files:           args.Files,
			RunOnlyCommands: args.RunOnlyCommands,
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		runner.RunAll(ctx, sourceDirs)
		close(resultChan)
	}()

	var results []run.Result
	for res := range resultChan {
		results = append(results, res)
	}

	if ctx.Err() != nil {
		return errors.New("Interrupted")
	}

	if !logSettings.SkipSummary() {
		printSummary(time.Since(startTime), results, logSettings)
	}

	for _, result := range results {
		if result.Status == run.StatusErr {
			return errors.New("") // No error should be printed
		}
	}

	return nil
}

func printSummary(
	duration time.Duration,
	results []run.Result,
	logSettings log.SkipSettings,
) {
	if len(results) == 0 {
		if !logSettings.SkipEmptySummary() {
			log.Separate(
				fmt.Sprintf(
					"%s %s %s",
					log.Cyan("summary:"),
					log.Gray("(skip)"),
					log.Yellow("empty"),
				),
			)
		}
		return
	}

	log.Separate(
		log.Cyan("summary: ") + log.Gray(fmt.Sprintf("(done in %.2f seconds)", duration.Seconds())),
	)

	if !logSettings.SkipSuccess() {
		for _, result := range results {
			if result.Status != run.StatusOk {
				continue
			}

			log.Infof("✔️  %s\n", log.Green(result.Name))
		}
	}

	if !logSettings.SkipFailure() {
		for _, result := range results {
			if result.Status != run.StatusErr {
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
