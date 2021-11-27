package config

import (
	"github.com/spf13/viper"

	"strings"

	"github.com/evilmartians/lefthook/pkg/log"
)

const (
	CMD = "{cmd}"
)

type Hook struct {
	// Should be unmarshalled with `mapstructure:"commands"`
	// But replacing '{cmd}' is still an issue
	// Unmarshaling it manually, so omit auto unmarshaling
	Commands map[string]*Command `mapstructure:"?"`

	// Should be unmarshalled with `mapstructure:"scripts"`
	// But parsing keys with dots in it is still an issue: https://github.com/spf13/viper/issues/324
	// Unmarshaling it manually, so omit auto unmarshaling
	Scripts map[string]*Script `mapstructure:"?"`

	Glob        string   `mapstructure:"glob"`
	Files       string   `mapstructure:"files"`
	Parallel    bool     `mapstructure:"parallel"`
	Piped       bool     `mapstructure:"piped"`
	ExcludeTags []string `mapstructure:"exclude_tags"`
}

func unmarshalHooks(base, extra *viper.Viper) (*Hook, error) {
	if base == nil && extra == nil {
		return nil, nil
	}

	commands, err := mergeCommands(base, extra)
	if err != nil {
		return nil, err
	}

	scripts, err := mergeScripts(base, extra)
	if err != nil {
		return nil, err
	}

	hook := Hook{
		Commands: commands,
		Scripts:  scripts,
	}

	if base == nil {
		base = extra
	} else if extra != nil {
		base.MergeConfigMap(extra.AllSettings())
	}

	if err := base.Unmarshal(&hook); err != nil {
		return nil, err
	}

	return &hook, nil
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

func replaceCmd(source, replacement string) string {
	return strings.Replace(source, CMD, replacement, -1)
}
