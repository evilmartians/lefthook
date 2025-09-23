package command

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
)

type JobParams struct {
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

type BuilderOptions struct {
	HookName   string
	GitArgs    []string
	ForceFiles []string
	SourceDirs []string
	Templates  map[string]string
	Force      bool
}

type Builder struct {
	git  *git.Repository
	opts BuilderOptions
}

func NewBuilder(repo *git.Repository, opts BuilderOptions) *Builder {
	return &Builder{
		git:  repo,
		opts: opts,
	}
}

// BuildCommands returns the list of commands and the list of files touched by the command.
func (b *Builder) BuildCommands(params *JobParams) ([]string, []string, error) {
	if len(params.Run) != 0 {
		return b.buildCommand(params)
	} else {
		return b.buildScript(params)
	}
}

func (p *JobParams) validateCommand() error {
	if !config.IsRunFilesCompatible(p.Run) {
		return config.ErrFilesIncompatible
	}

	return nil
}

func (p *JobParams) validateScript() error {
	return nil
}
