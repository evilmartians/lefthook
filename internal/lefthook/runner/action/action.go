package action

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
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

	Run       string
	Root      string
	Runner    string
	Script    string
	Glob      string
	Files     string
	FileTypes []string
	Tags      []string
	Exclude   interface{}
	Only      interface{}
	Skip      interface{}
}

type Action struct {
	Execs []string
	Files []string
}

func New(name string, params *Params) (*Action, error) {
	if params.skip() {
		return nil, SkipError{"settings"}
	}

	if intersect(params.Hook.ExcludeTags, params.Tags) {
		return nil, SkipError{"tags"}
	}

	if intersect(params.Hook.ExcludeTags, []string{name}) {
		return nil, SkipError{"name"}
	}

	var err error
	var action *Action
	if len(params.Run) != 0 {
		action, err = buildCommand(params)
	} else {
		action, err = buildScript(params)
	}

	if err != nil {
		return nil, err
	}

	return action, nil
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

func intersect(a, b []string) bool {
	intersections := make(map[string]struct{}, len(a))

	for _, v := range a {
		intersections[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := intersections[v]; ok {
			return true
		}
	}

	return false
}