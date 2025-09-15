package run

import (
	"maps"
	"slices"
	"sync/atomic"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/run/utils"
)

type scope struct {
	failed   *atomic.Bool
	parallel bool
	piped    bool
	follow   bool

	onlyJobs []string
	onlyTags []string

	glob        []string
	tags        []string
	excludeTags []string // Consider removing this setting
	names       []string
	exclude     interface{}
	env         map[string]string
	root        string
	hookName    string
	filesCmd    string
}

func (c *Controller) newScope(hookName string, hook *config.Hook) *scope {
	var failed atomic.Bool
	var exclude []interface{}
	if len(c.Exclude) > 0 {
		exclude = make([]interface{}, len(c.Exclude))
		for i, e := range c.Exclude {
			exclude[i] = e
		}
	}

	return &scope{
		hookName:    hookName,
		follow:      hook.Follow,
		failed:      &failed,
		filesCmd:    hook.Files,
		parallel:    hook.Parallel,
		piped:       hook.Piped,
		excludeTags: hook.ExcludeTags,
		exclude:     exclude,
		env:         make(map[string]string),
		onlyJobs:    c.RunOnlyJobs,
		onlyTags:    c.RunOnlyTags,
	}
}

func (s *scope) withOverwrites(job *config.Job) scope {
	newScope := *s
	newScope.parallel = job.Group.Parallel
	newScope.piped = job.Group.Piped
	newScope.glob = slices.Concat(newScope.glob, job.Glob)
	newScope.tags = slices.Concat(newScope.tags, job.Tags)
	newScope.root = utils.FirstNonBlank(job.Root, s.root)

	// Extend `exclude` list
	switch list := job.Exclude.(type) {
	case []interface{}:
		switch inherited := newScope.exclude.(type) {
		case []interface{}:
			// List of globs get appended
			inherited = append(inherited, list...)
			newScope.exclude = inherited
		default:
			// Regex value will be overwritten with a list of globs
			newScope.exclude = job.Exclude
		}
	case string:
		// Regex value always overwrites excludes
		newScope.exclude = job.Exclude
	default:
		// Inherit
	}

	// Overwrite --jobs option for nested groups: if group name given, run all its jobs
	if len(s.onlyJobs) != 0 && job.Group != nil && slices.Contains(s.onlyJobs, job.Name) {
		newScope.onlyJobs = []string{}
	}

	// Overwrite `files` command if present
	if len(job.Files) > 0 {
		s.filesCmd = job.Files
	}

	maps.Copy(newScope.env, job.Env)

	return newScope
}
