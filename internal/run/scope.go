package run

import "sync/atomic"

type scope struct {
	failed *atomic.Bool
	piped  bool

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
		piped:    c.Hook.Piped,
		onlyJobs: c.RunOnlyJobs,
		onlyTags: c.RunOnlyTags,
		exclude:  exclude,
		env:      make(map[string]string),
	}
}
