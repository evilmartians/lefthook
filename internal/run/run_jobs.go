package run

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/exec"
	"github.com/evilmartians/lefthook/internal/run/filters"
	"github.com/evilmartians/lefthook/internal/run/jobs"
	"github.com/evilmartians/lefthook/internal/run/result"
)

var (
	errJobContainsBothRunAndScript = errors.New("both `run` and `script` are not permitted")
	errEmptyJob                    = errors.New("no execution instructions")
	errEmptyGroup                  = errors.New("empty groups are not permitted")
)

type jobContext struct {
	failed *atomic.Bool

	onlyJobs []string
	onlyTags []string

	glob    []string
	tags    []string
	exclude interface{}
	names   []string
	env     map[string]string
	root    string
}

func newJobContext(exclude, onlyJobs, onlyTags []string) *jobContext {
	var failed atomic.Bool
	var excludeInterface []interface{}
	if len(exclude) > 0 {
		excludeInterface = make([]interface{}, len(exclude))
		for i, e := range exclude {
			excludeInterface[i] = e
		}
	}

	return &jobContext{
		failed:   &failed,
		onlyJobs: onlyJobs,
		onlyTags: onlyTags,
		exclude:  excludeInterface,
		env:      make(map[string]string),
	}
}

func (c *Controller) runJobs(ctx context.Context) []result.Result {
	var wg sync.WaitGroup

	results := make([]result.Result, 0, len(c.Hook.Jobs))
	resultsChan := make(chan result.Result, len(c.Hook.Jobs))

	jobContext := newJobContext(c.Exclude, c.RunOnlyJobs, c.RunOnlyTags)

	for i, job := range c.Hook.Jobs {
		id := strconv.Itoa(i)

		if jobContext.failed.Load() && c.Hook.Piped {
			log.Skip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !c.Hook.Parallel {
			results = append(results, c.runJob(ctx, jobContext, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- c.runJob(ctx, jobContext, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func (c *Controller) runJob(ctx context.Context, jobContext *jobContext, id string, job *config.Job) result.Result {
	startTime := time.Now()

	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return result.Failure(job.PrintableName(id), errJobContainsBothRunAndScript.Error(), time.Since(startTime))
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return result.Failure(job.PrintableName(id), errEmptyJob.Error(), time.Since(startTime))
	}

	if job.Interactive && !c.DisableTTY && !c.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		if len(jobContext.onlyJobs) != 0 && !slices.Contains(jobContext.onlyJobs, job.Name) {
			return result.Skip(job.PrintableName(id))
		}

		return c.runSingleJob(ctx, jobContext, id, job)
	}

	if job.Group != nil {
		inheritedJobContext := *jobContext
		inheritedJobContext.glob = slices.Concat(inheritedJobContext.glob, job.Glob)
		inheritedJobContext.tags = slices.Concat(inheritedJobContext.tags, job.Tags)
		inheritedJobContext.root = first(job.Root, jobContext.root)
		switch list := job.Exclude.(type) {
		case []interface{}:
			switch inherited := inheritedJobContext.exclude.(type) {
			case []interface{}:
				// List of globs get appended
				inherited = append(inherited, list...)
				inheritedJobContext.exclude = inherited
			default:
				// Regex value will be overwritten with a list of globs
				inheritedJobContext.exclude = job.Exclude
			}
		case string:
			// Regex value always overwrites excludes
			inheritedJobContext.exclude = job.Exclude
		default:
			// Inherit
		}
		groupName := first(job.Name, "group ("+id+")")
		inheritedJobContext.names = append(inheritedJobContext.names, groupName)

		if len(jobContext.onlyJobs) != 0 && slices.Contains(jobContext.onlyJobs, job.Name) {
			inheritedJobContext.onlyJobs = []string{}
		}

		maps.Copy(inheritedJobContext.env, job.Env)

		return c.runGroup(ctx, groupName, &inheritedJobContext, job.Group)
	}

	return result.Failure(job.PrintableName(id), "don't know how to run job", time.Since(startTime))
}

func (c *Controller) runSingleJob(ctx context.Context, jobContext *jobContext, id string, job *config.Job) result.Result {
	startTime := time.Now()

	name := job.PrintableName(id)

	root := first(job.Root, jobContext.root)
	glob := slices.Concat(jobContext.glob, job.Glob)
	exclude := join(job.Exclude, jobContext.exclude)
	tags := slices.Concat(job.Tags, jobContext.tags)
	executionJob, err := jobs.Build(&jobs.Params{
		Name:      name,
		Run:       job.Run,
		Root:      root,
		Runner:    job.Runner,
		Script:    job.Script,
		Glob:      glob,
		Files:     job.Files,
		FileTypes: job.FileTypes,
		Tags:      tags,
		Exclude:   exclude,
		Only:      job.Only,
		Skip:      job.Skip,
	}, &jobs.Settings{
		Repo:       c.Repo,
		Hook:       c.Hook,
		HookName:   c.HookName,
		ForceFiles: c.Files,
		Force:      c.Force,
		SourceDirs: c.SourceDirs,
		GitArgs:    c.GitArgs,
		OnlyTags:   jobContext.onlyTags,
		Templates:  c.Templates,
	})
	if err != nil {
		log.Skip(name, err.Error())

		var skipErr jobs.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(name)
		}

		jobContext.failed.Store(true)
		return result.Failure(name, err.Error(), time.Since(startTime))
	}

	env := maps.Clone(jobContext.env)
	maps.Copy(env, job.Env)
	ok := exec.Run(ctx, c.executor, &exec.RunOptions{
		Exec: exec.Options{
			Name:        strings.Join(append(jobContext.names, name), " ❯ "),
			Root:        filepath.Join(c.Repo.RootPath, root),
			Commands:    executionJob.Execs,
			Interactive: job.Interactive && !c.DisableTTY,
			UseStdin:    job.UseStdin,
			Env:         env,
		},
		Follow:      c.Hook.Follow,
		CachedStdin: c.cachedStdin,
	})

	executionTime := time.Since(startTime)

	if !ok {
		jobContext.failed.Store(true)
		return result.Failure(name, job.FailText, executionTime)
	}

	if config.HookUsesStagedFiles(c.HookName) && job.StageFixed {
		files := executionJob.Files

		if len(files) == 0 {
			var err error
			files, err = c.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result.Success(name, executionTime)
			}

			files = filters.Apply(c.Repo.Fs, files, filters.Params{
				Glob:      glob,
				Root:      root,
				Exclude:   exclude,
				FileTypes: job.FileTypes,
			})
		}

		if len(root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(root, file)
			}
		}

		c.addStagedFiles(files)
	}

	return result.Success(name, executionTime)
}

func (c *Controller) runGroup(ctx context.Context, groupName string, jobContext *jobContext, group *config.Group) result.Result {
	startTime := time.Now()

	if len(group.Jobs) == 0 {
		return result.Failure(groupName, errEmptyGroup.Error(), time.Since(startTime))
	}

	results := make([]result.Result, 0, len(group.Jobs))
	resultsChan := make(chan result.Result, len(group.Jobs))
	var wg sync.WaitGroup

	for i, job := range group.Jobs {
		id := strconv.Itoa(i)

		if jobContext.failed.Load() && group.Piped {
			log.Skip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !group.Parallel {
			results = append(results, c.runJob(ctx, jobContext, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- c.runJob(ctx, jobContext, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return result.Group(groupName, results)
}

func (c *Controller) addStagedFiles(files []string) {
	if err := c.Repo.AddFiles(files); err != nil {
		log.Warn("Couldn't stage fixed files:", err)
	}
}

// first finds first non-empty string and returns it.
func first(args ...string) string {
	for _, a := range args {
		if len(a) > 0 {
			return a
		}
	}

	return ""
}

func join(args ...interface{}) interface{} {
	result := []interface{}{}
	for _, a := range args {
		switch list := a.(type) {
		case []interface{}:
			result = append(result, list...)
		case interface{}:
			if len(result) > 0 {
				return result
			} else {
				return a
			}
		default:
		}
	}

	return result
}
