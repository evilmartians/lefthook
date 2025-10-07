package controller

import (
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
	excludeFiles []string
	env          map[string]string
	root         string
	hookName     string
	filesCmd     string
	opts         Options
}

func newScope(hook *config.Hook, opts Options) *scope {
	excludeFiles := make([]string, len(opts.ExcludeFiles)+len(hook.Exclude))

	i := 0
	for _, e := range opts.ExcludeFiles {
		excludeFiles[i] = e
		i += 1
	}
	for _, e := range hook.Exclude {
		excludeFiles[i] = e
		i += 1
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

	if len(job.Exclude) > 0 {
		newScope.excludeFiles = append(newScope.excludeFiles, job.Exclude...)
	}

	// Overwrite --job option for nested groups: if group name given, run all its jobs
	if len(s.opts.RunOnlyJobs) != 0 && job.Group != nil && slices.Contains(s.opts.RunOnlyJobs, job.Name) {
		newScope.opts.RunOnlyJobs = []string{}
	}

	// Copy env, avoid race conditions
	if len(job.Env) > 0 {
		if len(newScope.env) > 0 {
			env := make(map[string]string)
			for key, value := range newScope.env {
				env[key] = value
			}
			for key, value := range job.Env {
				env[key] = value
			}
			newScope.env = env
		} else {
			newScope.env = job.Env
		}
	}

	return &newScope
}
