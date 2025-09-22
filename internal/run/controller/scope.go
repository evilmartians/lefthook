package controller

import (
	"maps"
	"slices"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/run/controller/utils"
)

type scope struct {
	follow bool

	glob         []string
	tags         []string
	excludeTags  []string // Consider removing this setting
	names        []string
	fileTypes    []string
	excludeFiles interface{}
	env          map[string]string
	root         string
	hookName     string
	filesCmd     string
	opts         Options
}

func newScope(hook *config.Hook, opts Options) *scope {
	excludeFiles := make([]interface{}, len(opts.ExcludeFiles))
	if len(opts.ExcludeFiles) > 0 {
		for i, e := range opts.ExcludeFiles {
			excludeFiles[i] = e
		}
	}

	return &scope{
		hookName:     hook.Name,
		follow:       hook.Follow,
		filesCmd:     hook.Files,
		excludeTags:  hook.ExcludeTags,
		excludeFiles: excludeFiles,
		env:          make(map[string]string),
		opts:         opts,
	}
}

func (s *scope) extend(job *config.Job) *scope {
	newScope := *s
	newScope.glob = slices.Concat(newScope.glob, job.Glob)
	newScope.tags = slices.Concat(newScope.tags, job.Tags)
	newScope.root = utils.FirstNonBlank(job.Root, s.root)
	newScope.filesCmd = utils.FirstNonBlank(job.Files, s.filesCmd)
	newScope.fileTypes = slices.Concat(newScope.fileTypes, job.FileTypes)

	// Extend `exclude` list
	switch list := job.Exclude.(type) {
	case []interface{}:
		switch inherited := newScope.excludeFiles.(type) {
		case []interface{}:
			// List of globs get appended
			inherited = append(inherited, list...)
			newScope.excludeFiles = inherited
		default:
			// Regex value will be overwritten with a list of globs
			newScope.excludeFiles = job.Exclude
		}
	case string:
		// Regex value always overwrites excludes
		newScope.excludeFiles = job.Exclude
	default:
		// Inherit
	}

	// Overwrite --jobs option for nested groups: if group name given, run all its jobs
	if len(s.opts.RunOnlyJobs) != 0 && job.Group != nil && slices.Contains(s.opts.RunOnlyJobs, job.Name) {
		newScope.opts.RunOnlyJobs = []string{}
	}

	maps.Copy(newScope.env, job.Env)

	return &newScope
}
