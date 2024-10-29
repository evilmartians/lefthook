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
	errActionContainsBothRunAndScript = errors.New("both `run` and `script` are not permitted")
	errEmptyAction                    = errors.New("no execution instructions")
	errEmptyGroup                     = errors.New("empty groups are not permitted")
)

type domain struct {
	failed atomic.Bool
	glob   string
}

func (r *Runner) runActions(ctx context.Context) []Result {
	var wg sync.WaitGroup

	results := make([]Result, 0, len(r.Hook.Actions))
	resultsChan := make(chan Result, len(r.Hook.Actions))
	domain := &domain{}
	for i, action := range r.Hook.Actions {
		id := strconv.Itoa(i)

		if domain.failed.Load() && r.Hook.Piped {
			r.logSkip(action.PrintableName(id), "broken pipe")
			continue
		}

		if !r.Hook.Parallel {
			results = append(results, r.runAction(ctx, domain, id, action))
			continue
		}

		wg.Add(1)
		go func(action *config.Action) {
			defer wg.Done()
			resultsChan <- r.runAction(ctx, domain, id, action)
		}(action)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func (r *Runner) runAction(ctx context.Context, domain *domain, id string, action *config.Action) Result {
	// Check if do action is properly configured
	if len(action.Run) > 0 && len(action.Script) > 0 {
		return failed(action.PrintableName(id), errActionContainsBothRunAndScript.Error())
	}
	if len(action.Run) == 0 && len(action.Script) == 0 && action.Group == nil {
		return failed(action.PrintableName(id), errEmptyAction.Error())
	}

	if action.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	if len(action.Run) != 0 || len(action.Script) != 0 {
		return r.runSingleAction(ctx, domain, id, action)
	}

	if action.Group != nil {
		return r.runGroup(ctx, id, action.Group)
	}

	return failed(action.PrintableName(id), "don't know how to run action")
}

func (r *Runner) runSingleAction(ctx context.Context, domain *domain, id string, act *config.Action) Result {
	name := act.PrintableName(id)

	glob := act.Glob
	if len(glob) == 0 {
		glob = domain.glob
	}

	runAction, err := action.New(name, &action.Params{
		Repo:       r.Repo,
		Hook:       r.Hook,
		HookName:   r.HookName,
		ForceFiles: r.Files,
		Force:      r.Force,
		SourceDirs: r.SourceDirs,
		GitArgs:    r.GitArgs,
		Run:        act.Run,
		Root:       act.Root,
		Runner:     act.Runner,
		Script:     act.Script,
		Glob:       glob,
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

func (r *Runner) runGroup(ctx context.Context, groupId string, group *config.Group) Result {
	name := group.PrintableName(groupId)

	if len(group.Actions) == 0 {
		return failed(name, errEmptyGroup.Error())
	}

	results := make([]Result, 0, len(group.Actions))
	resultsChan := make(chan Result, len(group.Actions))
	domain := &domain{glob: group.Glob}
	var wg sync.WaitGroup

	for i, action := range group.Actions {
		id := strconv.Itoa(i)

		if domain.failed.Load() && group.Piped {
			r.logSkip(action.PrintableName(id), "broken pipe")
			continue
		}

		if !group.Parallel {
			results = append(results, r.runAction(ctx, domain, id, action))
			continue
		}

		wg.Add(1)
		go func(action *config.Action) {
			defer wg.Done()
			resultsChan <- r.runAction(ctx, domain, id, action)
		}(action)
	}

	wg.Wait()
	close(resultsChan)
	for result := range resultsChan {
		results = append(results, result)
	}

	return groupResult(name, results)
}
