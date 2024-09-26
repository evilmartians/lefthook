package config

import (
	"errors"
	"os"
	"strings"

	"github.com/knadh/koanf/v2"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

const CMD = "{cmd}"

var errPipedAndParallelSet = errors.New("conflicting options 'piped' and 'parallel' are set to 'true', remove one of this option from hook group")

type Hook struct {
	// Should be unmarshalled with `mapstructure:"commands"`
	// But replacing '{cmd}' is still an issue
	// Unmarshalling it manually, so omit auto unmarshalling
	Commands map[string]*Command `json:"commands,omitempty" mapstructure:"-" toml:"commands,omitempty" yaml:",omitempty"`

	// Should be unmarshalled with `mapstructure:"scripts"`
	// But parsing keys with dots in it is still an issue: https://github.com/spf13/viper/issues/324
	// Unmarshalling it manually, so omit auto unmarshalling
	Scripts map[string]*Script `json:"scripts,omitempty" mapstructure:"-" toml:"scripts,omitempty" yaml:",omitempty"`

	Files       string      `json:"files,omitempty"        mapstructure:"files"        toml:"files,omitempty"        yaml:",omitempty"`
	Parallel    bool        `json:"parallel,omitempty"     mapstructure:"parallel"     toml:"parallel,omitempty"     yaml:",omitempty"`
	Piped       bool        `json:"piped,omitempty"        mapstructure:"piped"        toml:"piped,omitempty"        yaml:",omitempty"`
	Follow      bool        `json:"follow,omitempty"       mapstructure:"follow"       toml:"follow,omitempty"       yaml:",omitempty"`
	ExcludeTags []string    `json:"exclude_tags,omitempty" mapstructure:"exclude_tags" toml:"exclude_tags,omitempty" yaml:"exclude_tags,omitempty"`
	Skip        interface{} `json:"skip,omitempty"         mapstructure:"skip"         toml:"skip,omitempty,inline"  yaml:",omitempty"`
	Only        interface{} `json:"only,omitempty"         mapstructure:"only"         toml:"only,omitempty,inline"  yaml:",omitempty"`
}

func (h *Hook) Validate() error {
	if h.Parallel && h.Piped {
		return errPipedAndParallelSet
	}

	return nil
}

func (h *Hook) DoSkip(gitState git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.check(gitState, h.Skip, h.Only)
}

func unmarshalHooks(base, extra *koanf.Koanf) (*Hook, error) {
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
		if err = base.Merge(extra); err != nil {
			return nil, err
		}
	}

	if err := base.Unmarshal("", &hook); err != nil {
		return nil, err
	}

	if tags := os.Getenv("LEFTHOOK_EXCLUDE"); tags != "" {
		hook.ExcludeTags = append(hook.ExcludeTags, strings.Split(tags, ",")...)
	}

	return &hook, nil
}
