package config

import (
	"errors"

	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/log"
)

const CMD = "{cmd}"

var errPipedAndParallelSet = errors.New("conflicting options 'piped' and 'parallel' are set to 'true', remove one of this option from hook group")

type Hook struct {
	// Should be unmarshalled with `mapstructure:"commands"`
	// But replacing '{cmd}' is still an issue
	// Unmarshaling it manually, so omit auto unmarshaling
	Commands map[string]*Command `mapstructure:"?"`

	// Should be unmarshalled with `mapstructure:"scripts"`
	// But parsing keys with dots in it is still an issue: https://github.com/spf13/viper/issues/324
	// Unmarshaling it manually, so omit auto unmarshaling
	Scripts map[string]*Script `mapstructure:"?"`

	Files       string   `mapstructure:"files"`
	Parallel    bool     `mapstructure:"parallel"`
	Piped       bool     `mapstructure:"piped"`
	ExcludeTags []string `mapstructure:"exclude_tags"`
}

func (h *Hook) Validate() error {
	if h.Parallel && h.Piped {
		return errPipedAndParallelSet
	}

	return nil
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
		if err = base.MergeConfigMap(extra.AllSettings()); err != nil {
			return nil, err
		}
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
