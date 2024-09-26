package lefthook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/runner"
	"github.com/evilmartians/lefthook/internal/log"
)

const (
	envEnabled    = "LEFTHOOK"        // "0", "false"
	envSkipOutput = "LEFTHOOK_QUIET"  // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
	envOutput     = "LEFTHOOK_OUTPUT" // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
)

type RunArgs struct {
	NoTTY           bool
	AllFiles        bool
	FilesFromStdin  bool
	Force           bool
	NoAutoInstall   bool
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
		var errNotFound config.ErrNotFound
		if ok := errors.As(err, &errNotFound); ok {
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
	var isGhostHook bool
	if hookName == config.GhostHookName && !ok && !verbose {
		isGhostHook = true
		log.SetLevel(log.WarnLevel)
	}

	enableLogTags := os.Getenv(envOutput)
	disableLogTags := os.Getenv(envSkipOutput)

	logSettings := log.NewSettings()
	logSettings.Apply(enableLogTags, disableLogTags, cfg.Output, cfg.SkipOutput)

	// Deprecate skip_output in the future. Leaving as is to reduce noise in output.
	// if outputSkipTags != "" || cfg.SkipOutput != nil {
	// 	 log.Warn("skip_output is deprecated, please use output option")
	// }

	if logSettings.LogMeta() {
		log.LogMeta(hookName)
	}

	if !args.NoAutoInstall {
		// This line controls updating the git hook if config has changed
		newCfg, err := l.syncHooks(cfg, !isGhostHook)
		if err != nil {
			log.Warnf(
				"⚠️  There was a problem with synchronizing git hooks. Run 'lefthook install' manually.\n   Error: %s", err,
			)
		} else {
			cfg = newCfg
		}
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

	if args.FilesFromStdin {
		paths, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read the files from standard input: %w", err)
		}
		args.Files = append(args.Files, parseFilesFromString(string(paths))...)
	} else if args.AllFiles {
		files, err := l.repo.AllFiles()
		if err != nil {
			return fmt.Errorf("failed to get all files: %w", err)
		}
		args.Files = append(args.Files, files...)
	}

	sourceDirs := []string{
		filepath.Join(l.repo.RootPath, cfg.SourceDir),
		filepath.Join(l.repo.RootPath, cfg.SourceDirLocal),
	}

	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			// Append only source_dir, because source_dir_local doesn't make sense
			sourceDirs = append(
				sourceDirs,
				filepath.Join(
					l.repo.RootPath,
					l.repo.RemoteFolder(remote.GitURL, remote.Ref),
					cfg.SourceDir,
				),
			)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	r := runner.New(
		runner.Options{
			Repo:            l.repo,
			Hook:            hook,
			HookName:        hookName,
			GitArgs:         gitArgs,
			LogSettings:     logSettings,
			DisableTTY:      cfg.NoTTY || args.NoTTY,
			Files:           args.Files,
			Force:           args.Force,
			RunOnlyCommands: args.RunOnlyCommands,
		},
	)

	startTime := time.Now()
	results, runErr := r.RunAll(ctx, sourceDirs)
	if runErr != nil {
		return fmt.Errorf("failed to run the hook: %w", runErr)
	}

	if ctx.Err() != nil {
		return errors.New("Interrupted")
	}

	printSummary(time.Since(startTime), results, logSettings)

	for _, result := range results {
		if result.Failure() {
			return errors.New("") // No error should be printed
		}
	}

	return nil
}

func printSummary(
	duration time.Duration,
	results []runner.Result,
	logSettings log.Settings,
) {
	if logSettings.LogSummary() {
		summaryPrint := log.Separate

		if !logSettings.LogExecution() {
			summaryPrint = func(s string) { log.Info(s) }
		}

		if len(results) == 0 {
			if logSettings.LogEmptySummary() {
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
	}

	if logSettings.LogSuccess() {
		for _, result := range results {
			if !result.Success() {
				continue
			}

			log.Success(result.Name)
		}
	}

	if logSettings.LogFailure() {
		for _, result := range results {
			if !result.Failure() {
				continue
			}

			log.Failure(result.Name, result.Text())
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

// parseFilesFromString parses both `\0`-separated files.
func parseFilesFromString(paths string) []string {
	var result []string
	start := 0
	for i, c := range paths {
		if c == 0 {
			result = append(result, paths[start:i])
			start = i + 1
		}
	}
	result = append(result, paths[start:])
	return result
}
