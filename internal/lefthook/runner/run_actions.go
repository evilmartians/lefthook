package runner

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/action"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/exec"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/filters"
	"github.com/evilmartians/lefthook/internal/log"
)

var (
	errJobContainsBothRunAndScript = errors.New("both `run` and `script` are not permitted")
	errEmptyJob                    = errors.New("no execution instructions")
	errEmptyGroup                  = errors.New("empty groups are not permitted")
)

type domain struct {
	failed *atomic.Bool

	glob string
	root string
}

func (r *Runner) runJobs(ctx context.Context) []Result {
	var wg sync.WaitGroup

	results := make([]Result, 0, len(r.Hook.Jobs))
	resultsChan := make(chan Result, len(r.Hook.Jobs))

	var failed atomic.Bool
	domain := &domain{failed: &failed}

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
	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return failed(job.PrintableName(id), errJobContainsBothRunAndScript.Error())
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return failed(job.PrintableName(id), errEmptyJob.Error())
	}

	if job.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		return r.runSingleJob(ctx, domain, id, job)
	}

	if job.Group != nil {
		inheritedDomain := *domain
		inheritedDomain.glob = first(job.Glob, domain.glob)
		inheritedDomain.root = first(job.Root, domain.root)
		groupName := first(job.Name, "["+id+"]")
		return r.runGroup(ctx, groupName, &inheritedDomain, job.Group)
	}

	return failed(job.PrintableName(id), "don't know how to run job")
}

func (r *Runner) runSingleJob(ctx context.Context, domain *domain, id string, act *config.Job) Result {
	name := act.PrintableName(id)

	runAction, err := action.New(name, &action.Params{
		Repo:       r.Repo,
		Hook:       r.Hook,
		HookName:   r.HookName,
		ForceFiles: r.Files,
		Force:      r.Force,
		SourceDirs: r.SourceDirs,
		GitArgs:    r.GitArgs,
		Run:        act.Run,
		Root:       first(act.Root, domain.root),
		Runner:     act.Runner,
		Script:     act.Script,
		Glob:       first(act.Glob, domain.glob),
		Files:      act.Files,
		FileTypes:  act.FileTypes,
		Tags:       act.Tags,
		Exclude:    act.Exclude,
		Only:       act.Only,
		Skip:       act.Skip,
	})
	if err != nil {
		r.logSkip(name, err.Error())

		var skipErr action.SkipError
		if errors.As(err, &skipErr) {
			return skipped(name)
		}

		domain.failed.Store(true)
		return failed(name, err.Error())
	}

	ok := r.run(ctx, exec.Options{
		Name:        name,
		Root:        filepath.Join(r.Repo.RootPath, act.Root),
		Commands:    runAction.Execs,
		Interactive: act.Interactive && !r.DisableTTY,
		UseStdin:    act.UseStdin,
		Env:         act.Env,
	}, r.Hook.Follow)

	if !ok {
		domain.failed.Store(true)
		return failed(name, act.FailText)
	}

	if config.HookUsesStagedFiles(r.HookName) && act.StageFixed {
		files := runAction.Files

		if len(files) == 0 {
			var err error
			files, err = r.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return succeeded(name)
			}

			files = filters.Apply(r.Repo.Fs, files, filters.Params{
				Glob:      act.Glob,
				Root:      act.Root,
				Exclude:   act.Exclude,
				FileTypes: act.FileTypes,
			})
		}

		if len(act.Root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(act.Root, file)
			}
		}

		r.addStagedFiles(files)
	}

	return succeeded(name)
}

func (r *Runner) runGroup(ctx context.Context, groupName string, domain *domain, group *config.Group) Result {
	if len(group.Jobs) == 0 {
		return failed(groupName, errEmptyGroup.Error())
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
