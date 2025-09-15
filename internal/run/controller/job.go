package controller

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/controller/filters"
	"github.com/evilmartians/lefthook/internal/run/controller/jobs"
	"github.com/evilmartians/lefthook/internal/run/controller/utils"
	"github.com/evilmartians/lefthook/internal/run/exec"
	"github.com/evilmartians/lefthook/internal/run/result"
)

const (
	invalidJobError = "either `run`,`script`, or `group` must be provided for a job"
	emptyGroupError = "group must have `jobs`"
)

func (c *Controller) runJob(ctx context.Context, scope *scope, id string, job *config.Job) result.Result {
	// Check if do job is properly configured
	if len(job.Run) > 0 && len(job.Script) > 0 {
		return result.Failure(job.PrintableName(id), invalidJobError, 0)
	}
	if len(job.Run) == 0 && len(job.Script) == 0 && job.Group == nil {
		return result.Failure(job.PrintableName(id), invalidJobError, 0)
	}

	startTime := time.Now()
	if job.Interactive && !scope.opts.DisableTTY && !scope.follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(job.Run) != 0 || len(job.Script) != 0 {
		if len(scope.opts.RunOnlyJobs) != 0 && !slices.Contains(scope.opts.RunOnlyJobs, job.Name) {
			return result.Skip(job.PrintableName(id))
		}

		return c.runSingleJob(ctx, scope, id, job)
	}

	if job.Group != nil {
		extendedScope := scope.extend(job)
		groupName := utils.FirstNonBlank(job.Name, "group ("+id+")")
		extendedScope.names = append(extendedScope.names, groupName)

		if len(job.Group.Jobs) == 0 {
			return result.Failure(groupName, emptyGroupError, 0)
		}

		var results []result.Result
		if job.Group.Parallel {
			results = c.concurrently(ctx, extendedScope, job.Group.Jobs)
		} else {
			results = c.sequentially(ctx, extendedScope, job.Group.Jobs, job.Group.Piped)
		}

		return result.Group(groupName, results)
	}

	return result.Failure(job.PrintableName(id), invalidJobError, time.Since(startTime))
}

func (c *Controller) runSingleJob(ctx context.Context, scope *scope, id string, job *config.Job) result.Result {
	startTime := time.Now()

	name := job.PrintableName(id)

	root := utils.FirstNonBlank(job.Root, scope.root)
	glob := slices.Concat(scope.glob, job.Glob)
	excludeFiles := joinInterfaces(job.Exclude, scope.excludeFiles)
	tags := slices.Concat(job.Tags, scope.tags)
	filesCmd := utils.FirstNonBlank(job.Files, scope.filesCmd)
	executionJob, err := jobs.Build(&jobs.Params{
		Name:         name,
		Run:          job.Run,
		Root:         root,
		Runner:       job.Runner,
		Script:       job.Script,
		Glob:         glob,
		FilesCmd:     filesCmd,
		FileTypes:    job.FileTypes,
		Tags:         tags,
		ExcludeFiles: excludeFiles,
		Only:         job.Only,
		Skip:         job.Skip,
	}, &jobs.Settings{
		Repo:        c.git,
		HookName:    scope.hookName,
		ExcludeTags: scope.excludeTags,
		ForceFiles:  scope.opts.Files,
		Force:       scope.opts.Force,
		SourceDirs:  scope.opts.SourceDirs,
		GitArgs:     scope.opts.GitArgs,
		OnlyTags:    scope.opts.RunOnlyTags,
		Templates:   scope.opts.Templates,
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
			Root:        filepath.Join(c.git.RootPath, root),
			Commands:    executionJob.Execs,
			Interactive: job.Interactive && !scope.opts.DisableTTY,
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
			files, err = c.git.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result.Success(name, executionTime)
			}

			files = filters.Apply(c.git.Fs, files, filters.Params{
				Glob:         glob,
				Root:         root,
				ExcludeFiles: excludeFiles,
				FileTypes:    job.FileTypes,
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
	if err := c.git.AddFiles(files); err != nil {
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
