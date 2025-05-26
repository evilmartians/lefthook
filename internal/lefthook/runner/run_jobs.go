package runner

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/exec"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/filters"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/jobs"
	"github.com/evilmartians/lefthook/internal/log"
)

var (
	errJobContainsBothRunAndScript = errors.New("both `run` and `script` are not permitted")
	errEmptyJob                    = errors.New("no execution instructions")
	errEmptyGroup                  = errors.New("empty groups are not permitted")
)

type domain struct {
	failed *atomic.Bool

	glob     []string
	root     string
	exclude  interface{}
	onlyJobs []string
	names    []string
}

func (r *Runner) runJobs(ctx context.Context) []Result {
	var wg sync.WaitGroup

	results := make([]Result, 0, len(r.Hook.Jobs))
	resultsChan := make(chan Result, len(r.Hook.Jobs))

	var failed atomic.Bool
	domain := &domain{failed: &failed, onlyJobs: r.RunOnlyJobs}

	for i, job := range r.Hook.Jobs {
		id := strconv.Itoa(i)

		if domain.failed.Load() && r.Hook.Piped {
			r.logSkip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !r.Hook.Parallel {
			results = append(results, r.runJob(ctx, domain, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- r.runJob(ctx, domain, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func (r *Runner) runJob(ctx context.Context, domain *domain, id string, job *config.Job) Result {
	startTime := time.Now()

	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return failed(job.PrintableName(id), errJobContainsBothRunAndScript.Error(), time.Since(startTime))
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return failed(job.PrintableName(id), errEmptyJob.Error(), time.Since(startTime))
	}

	if job.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		if len(domain.onlyJobs) != 0 && !slices.Contains(domain.onlyJobs, job.Name) {
			return skipped(job.PrintableName(id))
		}

		return r.runSingleJob(ctx, domain, id, job)
	}

	if job.Group != nil {
		inheritedDomain := *domain
		inheritedDomain.glob = slices.Concat(inheritedDomain.glob, job.Glob)
		inheritedDomain.root = first(job.Root, domain.root)
		switch list := job.Exclude.(type) {
		case []interface{}:
			switch inherited := inheritedDomain.exclude.(type) {
			case []interface{}:
				// List of globs get appended
				inherited = append(inherited, list...)
				inheritedDomain.exclude = inherited
			default:
				// Regex value will be overwritten with a list of globs
				inheritedDomain.exclude = job.Exclude
			}
		case string:
			// Regex value always overwrites excludes
			inheritedDomain.exclude = job.Exclude
		default:
			// Inherit
		}
		groupName := first(job.Name, "group ("+id+")")
		inheritedDomain.names = append(inheritedDomain.names, groupName)

		if len(domain.onlyJobs) != 0 && slices.Contains(domain.onlyJobs, job.Name) {
			inheritedDomain.onlyJobs = []string{}
		}

		return r.runGroup(ctx, groupName, &inheritedDomain, job.Group)
	}

	return failed(job.PrintableName(id), "don't know how to run job", time.Since(startTime))
}

func (r *Runner) runSingleJob(ctx context.Context, domain *domain, id string, job *config.Job) Result {
	startTime := time.Now()

	name := job.PrintableName(id)

	root := first(job.Root, domain.root)
	glob := slices.Concat(domain.glob, job.Glob)
	exclude := join(job.Exclude, domain.exclude)
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
			return skipped(name)
		}

		domain.failed.Store(true)
		return failed(name, err.Error(), time.Since(startTime))
	}

	ok := r.run(ctx, exec.Options{
		Name:        strings.Join(append(domain.names, name), " ❯ "),
		Root:        filepath.Join(r.Repo.RootPath, root),
		Commands:    executionJob.Execs,
		Interactive: job.Interactive && !r.DisableTTY,
		UseStdin:    job.UseStdin,
		Env:         job.Env,
	}, r.Hook.Follow)

	executionTime := time.Since(startTime)

	if !ok {
		domain.failed.Store(true)
		return failed(name, job.FailText, executionTime)
	}

	if config.HookUsesStagedFiles(r.HookName) && job.StageFixed {
		files := executionJob.Files

		if len(files) == 0 {
			var err error
			files, err = r.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return succeeded(name, executionTime)
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

	return succeeded(name, executionTime)
}

func (r *Runner) runGroup(ctx context.Context, groupName string, domain *domain, group *config.Group) Result {
	startTime := time.Now()

	if len(group.Jobs) == 0 {
		return failed(groupName, errEmptyGroup.Error(), time.Since(startTime))
	}

	results := make([]Result, 0, len(group.Jobs))
	resultsChan := make(chan Result, len(group.Jobs))
	var wg sync.WaitGroup

	for i, job := range group.Jobs {
		id := strconv.Itoa(i)

		if domain.failed.Load() && group.Piped {
			r.logSkip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !group.Parallel {
			results = append(results, r.runJob(ctx, domain, id, job))
			continue
		}

		wg.Add(1)
		go func(job *config.Job) {
			defer wg.Done()
			resultsChan <- r.runJob(ctx, domain, id, job)
		}(job)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return groupResult(groupName, results)
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
