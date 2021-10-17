package config

import (
	"strings"

	"github.com/evilmartians/lefthook/pkg/log"
)

type Hook struct {
	Commands map[string]*Command `mapstructure:"commands"`
	Scripts  map[string]*Script  `mapstructure:"scripts"`

	Glob  string `mapstructure:"glob"`
	Files string `mapstructure:"files"`

	Parallel bool `mapstructure:"parallel"`
	Piped    bool `mapstructure:"piped"`

	ExcludeTags []string `mapstructure:"exclude_tags"`
}

type Command struct {
	Run    string `mapstructure:"run"`
	Runner string `mapstructure:"runner"` // TODO: delete

	Skip bool     `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`

	Root    string   `mapstructure:"root"`
	Exclude []string `mapstructure:"exclude"`
}

type Script struct {
	Run    string `mapstructure:"run"` // TODO: delete
	Runner string `mapstructure:"runner"`

	Skip string   `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`
}

func (c *Command) RunValue() string {
	run := c.Run
	if run == "" && c.Runner != "" {
		log.Errorf("Warning: `runner` alias for commands is deprecated, use `run` instead.")
		run = c.Runner
	}
	return run
}

func (s *Script) RunnerValue() string {
	runner := s.Runner
	if runner == "" && s.Run != "" {
		log.Errorf("Warning: `run` alias for scripts is deprecated, use `runner` instead.")
		runner = s.Run
	}
	return runner
}

func (h *Hook) expandWith(baseHook *Hook) {
	for k, v := range h.Commands {

		run := v.RunValue()
		if res := strings.Contains(run, runnerWrapPattern); res {
			baseCmd := baseHook.Commands[k]

			if baseCmd != nil {
				run = strings.Replace(run, runnerWrapPattern, baseCmd.RunValue(), -1)
			}
		}

		v.Run = run
	}

	for k, v := range h.Scripts {
		runner := v.RunnerValue()

		if res := strings.Contains(runner, runnerWrapPattern); res {
			baseCmd := baseHook.Scripts[k]
			if baseCmd != nil {
				runner = strings.Replace(runner, runnerWrapPattern, baseCmd.RunnerValue(), -1)
			}
		}

		v.Runner = runner
	}
}
