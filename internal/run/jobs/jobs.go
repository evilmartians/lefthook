package jobs

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/run/utils"
	"github.com/evilmartians/lefthook/internal/system"
)

type Params struct {
	Name      string
	Run       string
	Root      string
	Runner    string
	Script    string
	Files     string
	FileTypes []string
	Tags      []string
	Glob      []string
	Exclude   interface{}
	Only      interface{}
	Skip      interface{}
}

type Settings struct {
	Repo       *git.Repository
	Hook       *config.Hook
	Force      bool
	HookName   string
	GitArgs    []string
	ForceFiles []string
	SourceDirs []string
	OnlyTags   []string
	Templates  map[string]string
}

type Job struct {
	Execs []string
	Files []string
}

func New(params *Params, settings *Settings) (*Job, error) {
	if settings.skip(params) {
		return nil, SkipError{"by condition"}
	}

	if utils.Intersect(settings.Hook.ExcludeTags, params.Tags) {
		return nil, SkipError{"tags"}
	}

	if len(settings.OnlyTags) > 0 && !utils.Intersect(settings.OnlyTags, params.Tags) {
		return nil, SkipError{"tags"}
	}

	if utils.Intersect(settings.Hook.ExcludeTags, []string{params.Name}) {
		return nil, SkipError{"name"}
	}

	var err error
	var job *Job
	if len(params.Run) != 0 {
		job, err = buildCommand(params, settings)
	} else {
		job, err = buildScript(params, settings)
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Settings) skip(params *Params) bool {
	skipChecker := config.NewSkipChecker(system.Cmd)
	return skipChecker.Check(s.Repo.State, params.Skip, params.Only)
}

func (p *Params) validateCommand() error {
	if !config.IsRunFilesCompatible(p.Run) {
		return config.ErrFilesIncompatible
	}

	return nil
}

func (p *Params) validateScript() error {
	return nil
}
