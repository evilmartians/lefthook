// Package runner handles ordering, filtering, substitutions while running
// jobs for a given hook.
package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/runner/executor"
	"github.com/evilmartians/lefthook/v2/internal/runner/result"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

type Runner struct {
	git          *git.Repo
	logger       *logger.ExecutionLogger
	stdin        func() io.Reader
	executor     executor.Executor
	cmd          system.CommandWithContext
	skipChecker  *config.SkipChecker
	filesToStage *stageFilesList
}

type Options struct {
	GitArgs           []string
	ExcludeFiles      []string
	Files             []string
	RunOnlyJobs       []string
	RunOnlyTags       []string
	SourceDirs        []string
	Templates         map[string]string
	GlobMatcher       string
	DisableTTY        bool
	FailOnChanges     bool
	FailOnChangesDiff bool
	Force             bool
	SkipLFS           bool
	NoStageFixed      bool
}

func New(repo *git.Repo, logger *logger.ExecutionLogger) *Runner {
	return &Runner{
		git:    repo,
		logger: logger,

		// Some hooks use STDIN for parsing data from Git. To allow multiple commands
		// and scripts access the same Git data STDIN is read once and shared.
		stdin: newStdin(os.Stdin, logger),

		// Executor interface for jobs
		executor: executor.New(logger),

		// Command interface (for LFS hooks)
		cmd: system.Cmd,

		skipChecker:  config.NewSkipChecker(logger, system.Cmd),
		filesToStage: newStageFilesList(),
	}
}

func (c *Runner) RunHook(ctx context.Context, opts Options, hook *config.Hook) ([]result.Result, error) {
	results := make([]result.Result, 0, len(hook.Jobs))

	if c.skipChecker.Check(c.git.State, hook.Skip, hook.Only) {
		c.logger.LogSkipped(hook.Name, "hook setting")
		return results, nil
	}

	if !opts.SkipLFS {
		if err := c.runLFSHook(ctx, hook.Name, opts.GitArgs); err != nil {
			return results, err
		}
	}

	if err := c.setup(ctx, opts, hook.Setup); err != nil {
		c.logger.Warnf("Failed to run setup: %s\n", err)
	}

	if !opts.DisableTTY && !hook.Follow {
		c.logger.Spinner.Start()
		defer c.logger.Spinner.Stop()
	}

	guard := newGuard(
		c.git,
		c.logger,
		c.filesToStage,
		!opts.NoStageFixed && config.HookUsesStagedFiles(hook.Name),
		opts.FailOnChanges,
		opts.FailOnChangesDiff,
	)
	scope := newScope(hook, opts)
	err := guard.wrap(func() {
		if hook.Parallel {
			results = c.concurrently(ctx, scope, hook.Jobs)
		} else {
			results = c.sequentially(ctx, scope, hook.Jobs, hook.Piped)
		}
	})
	if err != nil {
		return results, err
	}

	// If --job/--command was given but nothing in the hook matched it, every job
	// (and therefore every top-level result, since a Group's status is `skip` only
	// when all of its own sub-results are) was skipped rather than actually run.
	// Without this check, a typo'd job/command name silently "succeeds" having done
	// nothing at all.
	if len(opts.RunOnlyJobs) > 0 && len(results) > 0 {
		matched := false
		for _, res := range results {
			if !res.Skip() {
				matched = true
				break
			}
		}
		if !matched {
			return results, fmt.Errorf("no job matching %s found in hook %q", strings.Join(opts.RunOnlyJobs, ", "), hook.Name)
		}
	}

	return results, nil
}

func (c *Runner) concurrently(ctx context.Context, scope *scope, jobs []*config.Job) []result.Result {
	var wg sync.WaitGroup

	// Each job writes to its own index, so results keep the order of the config.
	results := make([]result.Result, len(jobs))
	for i, job := range jobs {
		wg.Go(func() {
			results[i] = c.runJob(ctx, scope, strconv.Itoa(i), job)
		})
	}
	wg.Wait()

	return results
}

func (c *Runner) sequentially(ctx context.Context, scope *scope, jobs []*config.Job, piped bool) []result.Result {
	results := make([]result.Result, 0, len(jobs))
	var failPipe bool

	for i, job := range jobs {
		id := strconv.Itoa(i)

		if piped && failPipe {
			c.logger.LogSkipped(job.PrintableName(id), "broken pipe")
			continue
		}

		result := c.runJob(ctx, scope, id, job)
		if piped && result.Failure() {
			failPipe = true
		}

		results = append(results, result)
	}

	return results
}
