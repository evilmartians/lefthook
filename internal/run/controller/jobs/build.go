package jobs

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/run/controller/utils"
	"github.com/evilmartians/lefthook/internal/system"
)

type Params struct {
	Name         string
	Run          string
	Root         string
	Runner       string
	Script       string
	FilesCmd     string
	FileTypes    []string
	Tags         []string
	Glob         []string
	ExcludeFiles interface{}
	Only         interface{}
	Skip         interface{}
}

type Settings struct {
	Repo        *git.Repository
	HookName    string
	ExcludeTags []string
	GitArgs     []string
	ForceFiles  []string
	SourceDirs  []string
	OnlyTags    []string
	Templates   map[string]string
	Force       bool
}

func Build(params *Params, settings *Settings) ([]string, []string, error) {
	if reason := settings.skipReason(params); len(reason) > 0 {
		return nil, nil, SkipError{reason}
	}

	if len(params.Run) != 0 {
		return buildCommand(params, settings)
	} else {
		return buildScript(params, settings)
	}
}

func (s *Settings) skipReason(params *Params) string {
	skipChecker := config.NewSkipChecker(system.Cmd)
	if skipChecker.Check(s.Repo.State, params.Skip, params.Only) {
		return "by condition"
	}

	if len(s.OnlyTags) > 0 && !utils.Intersect(s.OnlyTags, params.Tags) {
		return "tags"
	}

	if utils.Intersect(s.ExcludeTags, params.Tags) {
		return "tags"
	}

	if utils.Intersect(s.ExcludeTags, []string{params.Name}) {
		return "name"
	}

	return ""
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
