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

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run"
	"github.com/evilmartians/lefthook/v2/internal/run/result"
	"github.com/evilmartians/lefthook/v2/internal/version"
)

const (
	envEnabled = "LEFTHOOK"        // "0", "false"
	envOutput  = "LEFTHOOK_OUTPUT" // "meta,success,failure,summary,skips,execution,execution_out,execution_info"
)

var errPipedAndParallelSet = errors.New("conflicting options 'piped' and 'parallel' are set to 'true', remove one of this option from hook group")

type RunArgs struct {
	NoTTY           bool
	AllFiles        bool
	FilesFromStdin  bool
	Force           bool
	NoAutoInstall   bool
	NoStageFixed    bool
	SkipLFS         bool
	Verbose         bool
	FailOnChanges   *bool
	Hook            string
	Exclude         []string
	Files           []string
	RunOnlyCommands []string
	RunOnlyJobs     []string
	RunOnlyTags     []string
	GitArgs         []string
}

func (l *Lefthook) Run(ctx context.Context, args RunArgs) error {
	if os.Getenv(envEnabled) == "0" || os.Getenv(envEnabled) == "false" {
		return nil
	}

	waitPrecompute := l.repo.Precompute()
	defer waitPrecompute()

	if args.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Load config
	cfg, err := l.LoadConfig()
	if err != nil {
		var errNotFound config.ConfigNotFoundError
		if ok := errors.As(err, &errNotFound); ok {
			log.Warn(err.Error())
			return nil
		}
		return err
	}

	if err = checkVersion(cfg.MinVersion); err != nil {
		return err
	}

	// Suppress prepare-commit-msg output if the hook doesn't exist in config.
	// prepare-commit-msg hook is used for seamless synchronization of hooks with config.
	// See: internal/lefthook/install.go
	_, ok := cfg.Hooks[args.Hook]
	isGhostHook := args.Hook == config.GhostHookName && !ok && !args.Verbose
	if isGhostHook {
		log.SetLevel(log.WarnLevel)
	}

	enableLogTags := os.Getenv(envOutput)

	log.InitSettings()
	log.ApplySettings(enableLogTags, cfg.Output)

	if log.Settings.LogMeta() {
		log.LogMeta(args.Hook)
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

	hook, err := resolveHook(cfg, args.Hook)
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

	failOnChanges, err := shouldFailOnChanges(args.FailOnChanges, hook.FailOnChanges)
	if err != nil {
		return err
	}

	// Convert Commands and Scripts into Jobs
	hook.Jobs = append(hook.Jobs, config.CommandsToJobs(hook.Commands)...)
	hook.Commands = nil
	hook.Jobs = append(hook.Jobs, config.ScriptsToJobs(hook.Scripts)...)
	hook.Scripts = nil
	args.RunOnlyJobs = append(args.RunOnlyJobs, args.RunOnlyCommands...)

	return runHook(ctx, hook, l.repo, run.Options{
		DisableTTY:    cfg.NoTTY || args.NoTTY,
		SkipLFS:       cfg.SkipLFS || args.SkipLFS,
		Templates:     cfg.Templates,
		GitArgs:       args.GitArgs,
		ExcludeFiles:  args.Exclude,
		Files:         args.Files,
		Force:         args.Force,
		NoStageFixed:  args.NoStageFixed,
		RunOnlyJobs:   args.RunOnlyJobs,
		RunOnlyTags:   args.RunOnlyTags,
		SourceDirs:    sourceDirs,
		FailOnChanges: failOnChanges,
	})
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

func shouldFailOnChanges(fromArg *bool, fromHook string) (bool, error) {
	if fromArg != nil {
		return *fromArg, nil
	}

	switch fromHook {
	case "never", "false", "0", "":
		return false, nil
	case "always", "true", "1":
		return true, nil
	case "ci":
		_, ok := os.LookupEnv("CI")
		return ok, nil
	default:
		return false, fmt.Errorf("invalid value for fail_on_changes: %s", fromHook)
	}
}

func runHook(ctx context.Context, hook *config.Hook, repo *git.Repository, opts run.Options) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	startTime := time.Now()
	results, err := run.Run(ctx, hook, repo, opts)
	if err != nil {
		if errors.Is(err, run.ErrFailOnChanges) {
			return fmt.Errorf("%w", err)
		}
		return fmt.Errorf("failed to run the hook: %w", err)
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
				logResults(indent+1, result.Sub)
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

func checkVersion(minVersion string) error {
	if len(minVersion) == 0 {
		return nil
	}

	if err := version.Check(minVersion, version.Version(false)); err != nil {
		if errors.Is(err, version.ErrInvalidVersion) {
			return errors.New("format of 'min_version' setting is incorrect")
		}

		execPath, oserr := os.Executable()
		if oserr != nil {
			execPath = "<unknown>"
		}

		return fmt.Errorf("required lefthook version (%s) is higher than current (%s) at %s", minVersion, version.Version(false), execPath)
	}

	return nil
}
