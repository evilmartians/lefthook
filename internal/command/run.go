package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run"
	"github.com/evilmartians/lefthook/internal/run/result"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	envEnabled    = "LEFTHOOK"        // "0", "false"
	envSkipOutput = "LEFTHOOK_QUIET"  // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
	envOutput     = "LEFTHOOK_OUTPUT" // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
)

var errPipedAndParallelSet = errors.New("conflicting options 'piped' and 'parallel' are set to 'true', remove one of this option from hook group")

type RunArgs struct {
	NoTTY           bool
	AllFiles        bool
	FilesFromStdin  bool
	Force           bool
	NoAutoInstall   bool
	SkipLFS         bool
	Verbose         bool
	FailOnChanges   bool
	Exclude         []string
	Files           []string
	RunOnlyCommands []string
	RunOnlyJobs     []string
	RunOnlyTags     []string
}

func Run(opts *Options, args RunArgs, hookName string, gitArgs []string) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	args.Verbose = opts.Verbose

	return lefthook.Run(hookName, args, gitArgs)
}

func (l *Lefthook) Run(hookName string, args RunArgs, gitArgs []string) error {
	if os.Getenv(envEnabled) == "0" || os.Getenv(envEnabled) == "false" {
		return nil
	}

	waitPrecompute := l.repo.Precompute()
	defer waitPrecompute()

	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Load config
	cfg, err := config.Load(l.fs, l.repo)
	if err != nil {
		var errNotFound config.ConfigNotFoundError
		if ok := errors.As(err, &errNotFound); ok {
			log.Warn(err.Error())
			return nil
		}
		return err
	}

	if err = version.CheckCovered(cfg.MinVersion); err != nil {
		return err
	}

	// Suppress prepare-commit-msg output if the hook doesn't exist in config.
	// prepare-commit-msg hook is used for seamless synchronization of hooks with config.
	// See: internal/lefthook/install.go
	_, ok := cfg.Hooks[hookName]
	isGhostHook := hookName == config.GhostHookName && !ok && !args.Verbose
	if isGhostHook {
		log.SetLevel(log.WarnLevel)
	}

	enableLogTags := os.Getenv(envOutput)
	disableLogTags := os.Getenv(envSkipOutput)

	log.InitSettings()
	log.ApplySettings(enableLogTags, disableLogTags, cfg.Output, cfg.SkipOutput)

	// Deprecate skip_output in the future. Leaving as is to reduce noise in output.
	// if outputSkipTags != "" || cfg.SkipOutput != nil {
	// 	 log.Warn("skip_output is deprecated, please use output option")
	// }

	if log.Settings.LogMeta() {
		log.LogMeta(hookName)
	}

	if !args.NoAutoInstall {
		// This line controls updating the git hook if config has changed
		var newCfg *config.Config
		newCfg, err = l.syncHooks(cfg, !isGhostHook)
		if err != nil {
			log.Warnf(
				"⚠️  There was a problem with synchronizing git hooks. Run 'lefthook install' manually.\n   Error: %s", err,
			)
		} else {
			cfg = newCfg
		}
	}

	hook, err := resolveHook(cfg, hookName)
	if err != nil {
		return err
	}
	if hook == nil {
		return nil
	}

	files, err := getFiles(l.repo, args)
	if err != nil {
		return err
	}
	args.Files = files

	sourceDirs := getSourceDirs(l.repo, cfg)

	failOnChanges, err := shouldFailOnChanges(args, hook)
	if err != nil {
		return err
	}

	return executeHook(l.repo, cfg, hook, hookName, gitArgs, args, logSettings, sourceDirs, failOnChanges)
}

func resolveHook(cfg *config.Config, hookName string) (*config.Hook, error) {
	hook, ok := cfg.Hooks[hookName]
	if !ok {
		if config.KnownHook(hookName) {
			log.Debugf("[lefthook] skip: Hook %s doesn't exist in the config", hookName)
			return nil, nil
		}
		return nil, fmt.Errorf("hook %s doesn't exist in the config", hookName)
	}

	if hook.Parallel && hook.Piped {
		return nil, errPipedAndParallelSet
	}

	return hook, nil
}

func getFiles(repo *git.Repository, args RunArgs) ([]string, error) {
	if args.FilesFromStdin {
		paths, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read the files from standard input: %w", err)
		}
		return append(args.Files, parseFilesFromString(string(paths))...), nil
	} else if args.AllFiles {
		files, err := repo.AllFiles()
		if err != nil {
			return nil, fmt.Errorf("failed to get all files: %w", err)
		}
		return append(args.Files, files...), nil
	}

	return args.Files, nil
}

func getSourceDirs(repo *git.Repository, cfg *config.Config) []string {
	sourceDirs := []string{
		filepath.Join(repo.RootPath, cfg.SourceDir),
		filepath.Join(repo.RootPath, cfg.SourceDirLocal),

		// Additional source dirs to support .config/
		filepath.Join(repo.RootPath, ".config", "lefthook"),
		filepath.Join(repo.RootPath, ".config", "lefthook-local"),
	}

	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			// Append only source_dir, because source_dir_local doesn't make sense
			sourceDirs = append(
				sourceDirs,
				filepath.Join(
					repo.RemoteFolder(remote.GitURL, remote.Ref),
					cfg.SourceDir,
				),
			)
		}
	}

	return sourceDirs
}

func shouldFailOnChanges(args RunArgs, hook *config.Hook) (bool, error) {
	if args.FailOnChanges {
		return true, nil
	}

	switch hook.FailOnChanges {
	case "never", "":
		return false, nil
	case "always":
		return true, nil
	case "ci":
		_, ok := os.LookupEnv("CI")
		return ok, nil
	default:
		return false, fmt.Errorf("invalid value for fail_on_changes: %s", hook.FailOnChanges)
	}
}

func executeHook(
	repo *git.Repository, cfg *config.Config, hook *config.Hook, hookName string, gitArgs []string, args RunArgs,
	logSettings log.Settings, sourceDirs []string, failOnChanges bool,
) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	r := run.NewController(run.Options{
		Repo:            l.repo,
		Hook:            hook,
		HookName:        hookName,
		GitArgs:         gitArgs,
		DisableTTY:      cfg.NoTTY || args.NoTTY,
		SkipLFS:         cfg.SkipLFS || args.SkipLFS,
		Templates:       cfg.Templates,
		Exclude:         args.Exclude,
		Files:           args.Files,
		Force:           args.Force,
		RunOnlyCommands: args.RunOnlyCommands,
		RunOnlyJobs:     args.RunOnlyJobs,
		RunOnlyTags:     args.RunOnlyTags,
		SourceDirs:      sourceDirs,
		FailOnChanges:   failOnChanges,
	})

	startTime := time.Now()
	results, runErr := r.RunAll(ctx)
	if runErr != nil {
		if errors.Is(runErr, run.ErrFailOnChanges) {
			return fmt.Errorf("%w", runErr)
		}
		return fmt.Errorf("failed to run the hook: %w", runErr)
	}

	if ctx.Err() != nil {
		return errors.New("Interrupted")
	}

	printSummary(time.Since(startTime), results)

	for _, result := range results {
		if result.Failure() {
			return errors.New("") // No error should be printed
		}
	}

	return nil
}

func printSummary(
	duration time.Duration,
	results []result.Result,
) {
	if log.Settings.LogSummary() {
		summaryPrint := log.Separate

		if !log.Settings.LogExecution() {
			summaryPrint = func(s string) { log.Info(s) }
		}

		if len(results) == 0 {
			if log.Settings.LogEmptySummary() {
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

	logResults(0, results)
}

func logResults(indent int, results []result.Result) {
	if log.Settings.LogSuccess() {
		for _, result := range results {
			if !result.Success() {
				continue
			}

			log.Success(indent, result.Name, result.Duration)

			if len(result.Sub) > 0 {
				logResults(indent+1, result.Sub, logSettings)
			}
		}
	}

	if log.Settings.LogFailure() {
		for _, result := range results {
			if !result.Failure() {
				continue
			}

			log.Failure(indent, result.Name, result.Text(), result.Duration)

			if len(result.Sub) > 0 {
				logResults(indent+1, result.Sub)
			}
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
	cfg, err := config.Load(l.fs, l.repo)
	if err != nil {
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

func ConfigHookJobCompletions(opts *Options, hookName string) []string {
	lefthook, err := initialize(opts)
	if err != nil {
		return nil
	}
	return lefthook.configHookJobCompletions(hookName)
}

func (l *Lefthook) configHookCommandCompletions(hookName string) []string {
	cfg, err := config.Load(l.fs, l.repo)
	if err != nil {
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

func findJobNames(jobs []*config.Job) []string {
	jobNames := make([]string, 0, len(jobs))
	for _, job := range jobs {
		jobNames = append(jobNames, job.Name)
		if job.Group != nil {
			jobNames = append(jobNames, findJobNames(job.Group.Jobs)...)
		}
	}
	return jobNames
}

func (l *Lefthook) configHookJobCompletions(hookName string) []string {
	cfg, err := config.Load(l.fs, l.repo)
	if err != nil {
		return nil
	}
	if hook, found := cfg.Hooks[hookName]; !found {
		return nil
	} else {
		return findJobNames(hook.Jobs)
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
