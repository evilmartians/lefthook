package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"strings"

	"github.com/evilmartians/lefthook/pkg/log"
)

const (
	CMD = "{cmd}"
)

type Hook struct {
	Commands map[string]*Command `mapstructure:"commands"`

	// Should be unmarshalled with `mapstructure:"scripts"`
	// But still an issue: https://github.com/spf13/viper/issues/324
	// Unmarshaling it manually, so omit auto unmarshaling
	Scripts map[string]*Script `mapstructure:"?"`

	Glob  string `mapstructure:"glob"`
	Files string `mapstructure:"files"`

	Parallel        bool   `mapstructure:"parallel"`
	parallelDefined string `mapstructure:"parallel"`

	Piped        bool   `mapstructure:"piped"`
	pipedDefined string `mapstructure:"piped"`

	ExcludeTags []string `mapstructure:"exclude_tags"`
}

type Command struct {
	Run string `mapstructure:"run"`

	Skip []string `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`

	Root    string   `mapstructure:"root"`
	Exclude []string `mapstructure:"exclude"`

	// Deprecated
	Runner string `mapstructure:"runner"`
}

type Script struct {
	Runner string `mapstructure:"runner"`

	Skip []string `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`

	// Deprecated
	Run string `mapstructure:"run"`
}

func (c *Command) ToSkip() (skip bool, skipScripts []string) {
	skipScripts = c.Skip

	if len(c.Skip) == 1 {
		skip = c.Skip[0] == "1"
	}

	return
}

func (c *Command) Merge(another *Command) {
	if len(another.Skip) != 0 {
		c.Skip = another.Skip
	}

	if another.Root != "" {
		c.Root = another.Root
	}
	if len(another.Exclude) != 0 {
		c.Exclude = another.Exclude
	}
	if len(another.Tags) != 0 {
		c.Tags = another.Tags
	}

	runCmd := c.Run

	if strings.Contains(another.Run, CMD) {
		runCmd = strings.Replace(another.Run, CMD, runCmd, -1)
	} else if another.Run != "" {
		runCmd = another.Run
	}

	c.Run = runCmd
}

func (s *Script) ToSkip() (skip bool, skipScripts []string) {
	skipScripts = s.Skip

	if len(s.Skip) == 1 {
		skip = s.Skip[0] == "1"
	}

	return
}

func (s *Script) Merge(another *Script) {
	if len(another.Skip) != 0 {
		s.Skip = another.Skip
	}

	if len(another.Tags) != 0 {
		s.Tags = another.Tags
	}

	runnerCmd := s.Runner

	if strings.Contains(another.Runner, CMD) {
		runnerCmd = strings.Replace(another.Runner, CMD, runnerCmd, -1)
	} else if another.Runner != "" {
		runnerCmd = another.Runner
	}

	s.Runner = runnerCmd
}

// Merge another hook into current using special rules.
func (h *Hook) Merge(another *Hook) {
	if another.parallelDefined != "" {
		h.Parallel = another.Parallel
	}
	if another.pipedDefined != "" {
		h.Piped = another.Piped
	}

	if another.Glob != "" {
		h.Glob = another.Glob
	}
	if another.Files != "" {
		h.Files = another.Files
	}
	if len(another.ExcludeTags) != 0 {
		h.ExcludeTags = another.ExcludeTags
	}

	for name, command := range another.Commands {
		if h.Commands[name] == nil {
			h.Commands[name] = command
		} else {
			h.Commands[name].Merge(command)
		}
	}

	for name, script := range another.Scripts {
		if h.Scripts[name] == nil {
			h.Scripts[name] = script
		} else {
			h.Scripts[name].Merge(script)
		}
	}
}

func unmarshalHook(v *viper.Viper, hook *Hook) error {
	if err := v.Unmarshal(hook); err != nil {
		return err
	}

	// `scripts` are unmarshalled manually because mapstructure and viper
	// use "." as a key delimiter. So, this definition:
	//
	//   scripts:
	//     "example.sh":
	//       runner: bash
	//
	// Unmarshals into this:
	//
	//   scripts:
	//     example:
	//       sh:
	//         runner: bash
	//
	// This is not an expected behaviour and cannot be controlled yet
	// That's why we use manual unmarshaling using 'mapstructure'
	scripts := v.GetStringMap("scripts")

	if len(scripts) == 0 {
		return nil
	}

	hook.Scripts = make(map[string]*Script)

	for name, scriptConfig := range scripts {
		var script Script

		if err := mapstructure.WeakDecode(scriptConfig, &script); err != nil {
			return err
		}

		hook.Scripts[name] = &script
	}

	return nil
}

func (h Hook) processDeprecations() {
	var cmdDeprecationUsed, scriptDeprecationUsed bool

	for _, command := range h.Commands {
		if command.Runner != "" {
			cmdDeprecationUsed = true

			if command.Run == "" {
				command.Run = command.Runner
			}
		}
	}

	for _, script := range h.Scripts {
		if script.Run != "" {
			scriptDeprecationUsed = true

			if script.Runner == "" {
				script.Runner = script.Run
			}
		}
	}

	if cmdDeprecationUsed {
		log.Errorf("Warning: `runner` alias for commands is deprecated, use `run` instead.\n")
	}

	if scriptDeprecationUsed {
		log.Errorf("Warning: `run` alias for scripts is deprecated, use `runner` instead.\n")
	}
}
