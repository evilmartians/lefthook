// controller handles ordering, filtering, substitutions while running
// jobs for a given hook.
package controller

import (
	"context"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/utils"
	"github.com/evilmartians/lefthook/v2/internal/run/result"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

type Controller struct {
	git         *git.Repository
	cachedStdin io.Reader
	executor    exec.Executor
	cmd         system.CommandWithContext
}

type Options struct {
	GitArgs       []string
	ExcludeFiles  []string
	Files         []string
	RunOnlyJobs   []string
	RunOnlyTags   []string
	SourceDirs    []string
	Templates     map[string]string
	GlobMatcher   string
	DisableTTY    bool
	FailOnChanges bool
	Force         bool
	SkipLFS       bool
	NoStageFixed  bool
}

func NewController(repo *git.Repository) *Controller {
	return &Controller{
		git: repo,

		// Some hooks use STDIN for parsing data from Git. To allow multiple commands
		// and scripts access the same Git data STDIN is cached via CachedReadec.
		cachedStdin: utils.NewCachedReader(os.Stdin),

		// Executor interface for jobs
		executor: exec.CommandExecutor{},

		// Command interface (for LFS hooks)
		cmd: system.Cmd,
	}
}

func (c *Controller) RunHook(ctx context.Context, opts Options, hook *config.Hook) ([]result.Result, error) {
	results := make([]result.Result, 0, len(hook.Jobs))

	if config.NewSkipChecker(system.Cmd).Check(c.git.State, hook.Skip, hook.Only) {
		log.Skip(hook.Name, "hook setting")
		return results, nil
	}

	if !opts.SkipLFS {
		if err := c.runLFSHook(ctx, hook.Name, opts.GitArgs); err != nil {
			return results, err
		}
	}

	if !opts.DisableTTY && !hook.Follow {
		log.StartSpinner()
		defer log.StopSpinner()
	}

	guard := newGuard(c.git, !opts.NoStageFixed && config.HookUsesStagedFiles(hook.Name), opts.FailOnChanges)
	scope := newScope(hook, opts)
	err := guard.wrap(func() {
		if hook.Parallel {
			results = c.concurrently(ctx, scope, hook.Jobs)
		} else {
			results = c.sequentially(ctx, scope, hook.Jobs, hook.Piped)
		}
	})

	return results, err
}

func (c *Controller) concurrently(ctx context.Context, scope *scope, jobs []*config.Job) []result.Result {
	var wg sync.WaitGroup

	results := make([]result.Result, 0, len(jobs))
	resultsChan := make(chan result.Result, len(jobs))

	for i, job := range jobs {
		id := strconv.Itoa(i)

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- c.runJob(ctx, scope, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func (c *Controller) sequentially(ctx context.Context, scope *scope, jobs []*config.Job, piped bool) []result.Result {
	results := make([]result.Result, 0, len(jobs))
	var failPipe bool

	for i, job := range jobs {
		id := strconv.Itoa(i)

		if piped && failPipe {
			log.Skip(job.PrintableName(id), "broken pipe")
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
