package lefthook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
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
	Force           bool
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

	// Suppress prepare-commit-msg output if the hook doesn't exist in config.
	// prepare-commit-msg hook is used for seamless synchronization of hooks with config.
	// See: internal/lefthook/install.go
	_, ok := cfg.Hooks[hookName]
	if hookName == config.GhostHookName && !ok && !verbose {
		log.SetLevel(log.WarnLevel)
	}

	tags := os.Getenv(envSkipOutput)

	var logSettings log.SkipSettings
	(&logSettings).ApplySettings(tags, cfg.SkipOutput)

	if !logSettings.SkipMeta() {
		log.Box(
			log.Cyan("🥊 lefthook ")+log.Gray(fmt.Sprintf("v%s", version.Version(false))),
			log.Gray("hook: ")+log.Bold(hookName),
		)
	}

	// This line controls updating the git hook if config has changed
	if err = l.createHooksIfNeeded(cfg, true, false); err != nil {
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
			Force:           args.Force,
			RunOnlyCommands: args.RunOnlyCommands,
		},
	)

	sourceDirs := []string{
		filepath.Join(l.repo.RootPath, cfg.SourceDir),
		filepath.Join(l.repo.RootPath, cfg.SourceDirLocal),
	}

	// For backward compatibility with single remote config
	if cfg.Remote != nil {
		cfg.Remotes = append(cfg.Remotes, cfg.Remote)
	}

	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			// Append only source_dir, because source_dir_local doesn't make sense
			sourceDirs = append(
				sourceDirs,
				filepath.Join(
					l.repo.RemoteFolder(remote.GitURL, remote.Ref),
					cfg.SourceDir,
				),
			)
		}
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
	summaryPrint := log.Separate

	if logSettings.SkipExecution() || (logSettings.SkipExecutionInfo() && logSettings.SkipExecutionOutput()) {
		summaryPrint = func(s string) { log.Info(s) }
	}

	if len(results) == 0 {
		if !logSettings.SkipEmptySummary() {
			summaryPrint(
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

	summaryPrint(
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

func ConfigHookCompletions(opts *Options) []string {
	lefthook, err := initialize(opts)
	if err != nil {
		return nil
	}
	return lefthook.configHookCompletions()
}

func (l *Lefthook) configHookCompletions() []string {
	cfg, err := config.Load(l.Fs, l.repo)
	if err != nil {
		return nil
	}
	if err = cfg.Validate(); err != nil {
		return nil
	}
	hooks := make([]string, 0, len(cfg.Hooks))
	for hook := range cfg.Hooks {
		hooks = append(hooks, hook)
	}
	return hooks
}

func ConfigHookCommandCompletions(opts *Options, hookName string) []string {
	lefthook, err := initialize(opts)
	if err != nil {
		return nil
	}
	return lefthook.configHookCommandCompletions(hookName)
}

func (l *Lefthook) configHookCommandCompletions(hookName string) []string {
	cfg, err := config.Load(l.Fs, l.repo)
	if err != nil {
		return nil
	}
	if err = cfg.Validate(); err != nil {
		return nil
	}
	if hook, found := cfg.Hooks[hookName]; !found {
		return nil
	} else {
		commands := make([]string, 0, len(hook.Commands))
		for command := range hook.Commands {
			commands = append(commands, command)
		}
		return commands
	}
}
