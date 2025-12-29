package controller

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/command"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/utils"
	"github.com/evilmartians/lefthook/v2/internal/run/result"
	"github.com/evilmartians/lefthook/v2/internal/system"
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

		if len(scope.opts.RunOnlyTags) != 0 && (!utils.Intersect(scope.opts.RunOnlyTags, job.Tags) && !utils.Intersect(scope.opts.RunOnlyTags, scope.tags)) {
			return result.Skip(job.PrintableName(id))
		}

		return c.runSingleJob(ctx, scope, id, job)
	}

	if job.Group != nil {
		extendedScope := scope.extend(job)
		groupName := utils.FirstNonBlank(job.Name, "group ("+id+")")

		if reason := c.skipReason(extendedScope, job, groupName); len(reason) > 0 {
			log.Skip(groupName, reason)

			return result.Skip(groupName)
		}

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
	scope = scope.extend(job)

	if reason := c.skipReason(scope, job, name); len(reason) > 0 {
		log.Skip(name, reason)

		return result.Skip(name)
	}

	builder := command.NewBuilder(c.git, command.BuilderOptions{
		HookName:    scope.hookName,
		ForceFiles:  scope.opts.Files,
		Force:       scope.opts.Force,
		SourceDirs:  scope.opts.SourceDirs,
		GitArgs:     scope.opts.GitArgs,
		Templates:   scope.opts.Templates,
		GlobMatcher: scope.opts.GlobMatcher,
	})
	commands, files, err := builder.BuildCommands(&command.JobParams{
		Name:         name,
		Run:          job.Run,
		Runner:       job.Runner,
		Args:         job.Args,
		Script:       job.Script,
		Only:         job.Only,
		Skip:         job.Skip,
		Root:         scope.root,
		FileTypes:    scope.fileTypes,
		Glob:         scope.glob,
		FilesCmd:     scope.filesCmd,
		Tags:         scope.tags,
		ExcludeFiles: scope.excludeFiles,
	})
	if err != nil {
		log.Skip(name, err.Error())

		var skipErr command.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(name)
		}

		return result.Failure(name, err.Error(), time.Since(startTime))
	}

	env := maps.Clone(scope.env)
	maps.Copy(env, job.Env)
	ok := c.run(ctx, strings.Join(append(scope.names, name), " ❯ "), scope.follow, exec.Options{
		Root:        filepath.Join(c.git.RootPath, scope.root),
		Commands:    commands,
		Interactive: job.Interactive && !scope.opts.DisableTTY,
		UseStdin:    job.UseStdin,
		Timeout:     job.Timeout,
		Env:         env,
	})

	executionTime := time.Since(startTime)

	if !ok {
		return result.Failure(name, job.FailText, executionTime)
	}

	if config.HookUsesStagedFiles(scope.hookName) && job.StageFixed && !scope.opts.NoStageFixed {
		if len(files) == 0 {
			var err error
			files, err = c.git.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result.Success(name, executionTime)
			}

			files = filter.New(c.git.Fs, filter.Params{
				Glob:         scope.glob,
				Root:         scope.root,
				ExcludeFiles: scope.excludeFiles,
				FileTypes:    scope.fileTypes,
				GlobMatcher:  scope.opts.GlobMatcher,
			}).Apply(files)
		}

		if len(scope.root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(scope.root, file)
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

func (c *Controller) skipReason(scope *scope, job *config.Job, name string) string {
	skipChecker := config.NewSkipChecker(system.Cmd)
	if skipChecker.Check(c.git.State, job.Skip, job.Only) {
		return "by condition"
	}

	if utils.Intersect(scope.excludeTags, scope.tags) {
		return "tags"
	}

	if utils.Intersect(scope.excludeTags, []string{name}) {
		return "name"
	}

	return ""
}
