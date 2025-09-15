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
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/exec"
	"github.com/evilmartians/lefthook/internal/run/filters"
	"github.com/evilmartians/lefthook/internal/run/jobs"
	"github.com/evilmartians/lefthook/internal/run/result"
	"github.com/evilmartians/lefthook/internal/run/utils"
)

var (
	errJobContainsBothRunAndScript = errors.New("both `run` and `script` are not permitted")
	errEmptyJob                    = errors.New("no execution instructions")
	errEmptyGroup                  = errors.New("empty groups are not permitted")
)

func (c *Controller) runJobs(ctx context.Context, scope *scope, jobs []*config.Job) []result.Result {
	var wg sync.WaitGroup

	results := make([]result.Result, 0, len(jobs))
	resultsChan := make(chan result.Result, len(jobs))

	for i, job := range jobs {
		id := strconv.Itoa(i)

		if scope.failed.Load() && scope.piped {
			log.Skip(job.PrintableName(id), "broken pipe")
			continue
		}

		if !scope.parallel {
			results = append(results, c.runJob(ctx, scope, id, job))
			continue
		}

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

func (c *Controller) runJob(ctx context.Context, scope *scope, id string, job *config.Job) result.Result {
	startTime := time.Now()

	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return result.Failure(job.PrintableName(id), errJobContainsBothRunAndScript.Error(), time.Since(startTime))
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return result.Failure(job.PrintableName(id), errEmptyJob.Error(), time.Since(startTime))
	}

	if job.Interactive && !c.DisableTTY && !scope.follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		if len(scope.onlyJobs) != 0 && !slices.Contains(scope.onlyJobs, job.Name) {
			return result.Skip(job.PrintableName(id))
		}

		return c.runSingleJob(ctx, scope, id, job)
	}

	if job.Group != nil {
		inheritedScope := scope.withOverwrites(job)
		groupName := utils.FirstNonBlank(job.Name, "group ("+id+")")
		inheritedScope.names = append(inheritedScope.names, groupName)

		return c.runGroup(ctx, groupName, &inheritedScope, job.Group)
	}

	return result.Failure(job.PrintableName(id), "don't know how to run job", time.Since(startTime))
}

func (c *Controller) runGroup(ctx context.Context, groupName string, scope *scope, group *config.Group) result.Result {
	startTime := time.Now()

	if len(group.Jobs) == 0 {
		return result.Failure(groupName, errEmptyGroup.Error(), time.Since(startTime))
	}

	results := c.runJobs(ctx, scope, group.Jobs)

	return result.Group(groupName, results)
}

func (c *Controller) runSingleJob(ctx context.Context, scope *scope, id string, job *config.Job) result.Result {
	startTime := time.Now()

	name := job.PrintableName(id)

	root := utils.FirstNonBlank(job.Root, scope.root)
	glob := slices.Concat(scope.glob, job.Glob)
	exclude := joinInterfaces(job.Exclude, scope.exclude)
	tags := slices.Concat(job.Tags, scope.tags)
	executionJob, err := jobs.Build(&jobs.Params{
		Name:      name,
		Run:       job.Run,
		Root:      root,
		Runner:    job.Runner,
		Script:    job.Script,
		Glob:      glob,
		FilesCmd:  scope.filesCmd,
		FileTypes: job.FileTypes,
		Tags:      tags,
		Exclude:   exclude,
		Only:      job.Only,
		Skip:      job.Skip,
	}, &jobs.Settings{
		Repo:        c.Repo,
		HookName:    scope.hookName,
		ExcludeTags: scope.excludeTags,
		ForceFiles:  c.Files,
		Force:       c.Force,
		SourceDirs:  c.SourceDirs,
		GitArgs:     c.GitArgs,
		OnlyTags:    scope.onlyTags,
		Templates:   c.Templates,
	})
	if err != nil {
		log.Skip(name, err.Error())

		var skipErr jobs.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(name)
		}

		scope.failed.Store(true)
		return result.Failure(name, err.Error(), time.Since(startTime))
	}

	env := maps.Clone(scope.env)
	maps.Copy(env, job.Env)
	ok := exec.Run(ctx, c.executor, &exec.RunOptions{
		Exec: exec.Options{
			Name:        strings.Join(append(scope.names, name), " ❯ "),
			Root:        filepath.Join(c.Repo.RootPath, root),
			Commands:    executionJob.Execs,
			Interactive: job.Interactive && !c.DisableTTY,
			UseStdin:    job.UseStdin,
			Env:         env,
		},
		Follow:      scope.follow,
		CachedStdin: c.cachedStdin,
	})

	executionTime := time.Since(startTime)

	if !ok {
		scope.failed.Store(true)
		return result.Failure(name, job.FailText, executionTime)
	}

	if config.HookUsesStagedFiles(scope.hookName) && job.StageFixed {
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

func (c *Controller) addStagedFiles(files []string) {
	if err := c.Repo.AddFiles(files); err != nil {
		log.Warn("Couldn't stage fixed files:", err)
	}
}

func joinInterfaces(args ...interface{}) interface{} {
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
