package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
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

	Files       string      `mapstructure:"files"`
	Parallel    bool        `mapstructure:"parallel"`
	Piped       bool        `mapstructure:"piped"`
	ExcludeTags []string    `mapstructure:"exclude_tags"`
	Skip        interface{} `mapstructure:"skip"`
	Follow      bool        `mapstructure:"follow"`
}

func (h *Hook) Validate() error {
	if h.Parallel && h.Piped {
		return errPipedAndParallelSet
	}

	return nil
}

func (h *Hook) DoSkip(gitState git.State) bool {
	if value := h.Skip; value != nil {
		return isSkip(gitState, value)
	}
	return false
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

	if tags := os.Getenv("LEFTHOOK_EXCLUDE"); tags != "" {
		hook.ExcludeTags = append(hook.ExcludeTags, strings.Split(tags, ",")...)
	}

	return &hook, nil
}
