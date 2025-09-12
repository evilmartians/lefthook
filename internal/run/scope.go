package run

import (
	"maps"
	"slices"
	"sync/atomic"

	"github.com/evilmartians/lefthook/internal/config"
)

type scope struct {
	failed   *atomic.Bool
	parallel bool
	piped    bool

	onlyJobs []string
	onlyTags []string

	glob    []string
	tags    []string
	exclude interface{}
	names   []string
	env     map[string]string
	root    string
}

func (c *Controller) newScope() *scope {
	var failed atomic.Bool
	var exclude []interface{}
	if len(c.Exclude) > 0 {
		exclude = make([]interface{}, len(c.Exclude))
		for i, e := range c.Exclude {
			exclude[i] = e
		}
	}

	return &scope{
		failed:   &failed,
		parallel: c.Hook.Parallel,
		piped:    c.Hook.Piped,
		onlyJobs: c.RunOnlyJobs,
		onlyTags: c.RunOnlyTags,
		exclude:  exclude,
		env:      make(map[string]string),
	}
}

func (s *scope) withOverwrites(job *config.Job) scope {
	newScope := *s
	newScope.parallel = job.Group.Parallel
	newScope.piped = job.Group.Piped
	newScope.glob = slices.Concat(newScope.glob, job.Glob)
	newScope.tags = slices.Concat(newScope.tags, job.Tags)
	newScope.root = first(job.Root, s.root)
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
	if len(s.onlyJobs) != 0 && slices.Contains(s.onlyJobs, job.Name) {
		newScope.onlyJobs = []string{}
	}

	maps.Copy(newScope.env, job.Env)

	return newScope
}
