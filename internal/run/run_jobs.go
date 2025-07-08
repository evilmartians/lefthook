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

	glob     []string
	root     string
	exclude  interface{}
	onlyJobs []string
	names    []string
	env      map[string]string
}

func newJobContext(onlyJobs []string, exclude []string) *jobContext {
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
		exclude:  excludeInterface,
		env:      make(map[string]string),
	}
}

func (r *Run) runJobs(ctx context.Context) []result.Result {
	var wg sync.WaitGroup

	results := make([]result.Result, 0, len(r.Hook.Jobs))
	resultsChan := make(chan result.Result, len(r.Hook.Jobs))

	jobContext := newJobContext(r.RunOnlyJobs, r.Exclude)

	for i, job := range r.Hook.Jobs {
		id := strconv.Itoa(i)

		if jobContext.failed.Load() && r.Hook.Piped {
			r.logSkip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !r.Hook.Parallel {
			results = append(results, r.runJob(ctx, jobContext, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- r.runJob(ctx, jobContext, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func (r *Run) runJob(ctx context.Context, jobContext *jobContext, id string, job *config.Job) result.Result {
	startTime := time.Now()

	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return result.Failure(job.PrintableName(id), errJobContainsBothRunAndScript.Error(), time.Since(startTime))
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return result.Failure(job.PrintableName(id), errEmptyJob.Error(), time.Since(startTime))
	}

	if job.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		if len(jobContext.onlyJobs) != 0 && !slices.Contains(jobContext.onlyJobs, job.Name) {
			return result.Skip(job.PrintableName(id))
		}

		return r.runSingleJob(ctx, jobContext, id, job)
	}

	if job.Group != nil {
		inheritedJobContext := *jobContext
		inheritedJobContext.glob = slices.Concat(inheritedJobContext.glob, job.Glob)
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

		return r.runGroup(ctx, groupName, &inheritedJobContext, job.Group)
	}

	return result.Failure(job.PrintableName(id), "don't know how to run job", time.Since(startTime))
}

func (r *Run) runSingleJob(ctx context.Context, jobContext *jobContext, id string, job *config.Job) result.Result {
	startTime := time.Now()

	name := job.PrintableName(id)

	root := first(job.Root, jobContext.root)
	glob := slices.Concat(jobContext.glob, job.Glob)
	exclude := join(job.Exclude, jobContext.exclude)
	executionJob, err := jobs.New(name, &jobs.Params{
		Repo:       r.Repo,
		Hook:       r.Hook,
		HookName:   r.HookName,
		ForceFiles: r.Files,
		Force:      r.Force,
		SourceDirs: r.SourceDirs,
		GitArgs:    r.GitArgs,
		Run:        job.Run,
		Root:       root,
		Runner:     job.Runner,
		Script:     job.Script,
		Glob:       glob,
		Files:      job.Files,
		FileTypes:  job.FileTypes,
		Tags:       job.Tags,
		Exclude:    exclude,
		Only:       job.Only,
		Skip:       job.Skip,
		Templates:  r.Templates,
	})
	if err != nil {
		r.logSkip(name, err.Error())

		var skipErr jobs.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(name)
		}

		jobContext.failed.Store(true)
		return result.Failure(name, err.Error(), time.Since(startTime))
	}

	env := maps.Clone(jobContext.env)
	maps.Copy(env, job.Env)
	ok := r.run(ctx, exec.Options{
		Name:        strings.Join(append(jobContext.names, name), " ❯ "),
		Root:        filepath.Join(r.Repo.RootPath, root),
		Commands:    executionJob.Execs,
		Interactive: job.Interactive && !r.DisableTTY,
		UseStdin:    job.UseStdin,
		Env:         env,
	}, r.Hook.Follow)

	executionTime := time.Since(startTime)

	if !ok {
		jobContext.failed.Store(true)
		return result.Failure(name, job.FailText, executionTime)
	}

	if config.HookUsesStagedFiles(r.HookName) && job.StageFixed {
		files := executionJob.Files

		if len(files) == 0 {
			var err error
			files, err = r.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result.Success(name, executionTime)
			}

			files = filters.Apply(r.Repo.Fs, files, filters.Params{
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

		r.addStagedFiles(files)
	}

	return result.Success(name, executionTime)
}

func (r *Run) runGroup(ctx context.Context, groupName string, jobContext *jobContext, group *config.Group) result.Result {
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
			r.logSkip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !group.Parallel {
			results = append(results, r.runJob(ctx, jobContext, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- r.runJob(ctx, jobContext, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return result.Group(groupName, results)
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
