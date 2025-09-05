package jobs

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/run/utils"
	"github.com/evilmartians/lefthook/internal/system"
)

type Params struct {
	Repo       *git.Repository
	Hook       *config.Hook
	HookName   string
	GitArgs    []string
	Force      bool
	ForceFiles []string
	SourceDirs []string

	Name      string
	Run       string
	Root      string
	Runner    string
	Script    string
	Files     string
	FileTypes []string
	Tags      []string
	Glob      []string
	Templates map[string]string
	Exclude   interface{}
	Only      interface{}
	Skip      interface{}
}

type Job struct {
	Execs []string
	Files []string
}

func New(name string, params *Params) (*Job, error) {
	if params.skip() {
		return nil, SkipError{"by condition"}
	}

	if utils.Intersect(params.Hook.ExcludeTags, params.Tags) {
		return nil, SkipError{"tags"}
	}

	if utils.Intersect(params.Hook.ExcludeTags, []string{name}) {
		return nil, SkipError{"name"}
	}

	var err error
	var job *Job
	if len(params.Run) != 0 {
		job, err = buildCommand(params)
	} else {
		job, err = buildScript(params)
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (p *Params) skip() bool {
	skipChecker := config.NewSkipChecker(system.Cmd)
	return skipChecker.Check(p.Repo.State, p.Skip, p.Only)
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
