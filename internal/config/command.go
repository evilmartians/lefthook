package config

import (
	"errors"
	"strings"

	"github.com/knadh/koanf/v2"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

var errFilesIncompatible = errors.New("One of your runners contains incompatible file types")

type Command struct {
	Run   string `json:"run"             mapstructure:"run"   toml:"run"             yaml:"run"`
	Files string `json:"files,omitempty" mapstructure:"files" toml:"files,omitempty" yaml:",omitempty"`

	Skip interface{}       `json:"skip,omitempty" mapstructure:"skip" toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only interface{}       `json:"only,omitempty" mapstructure:"only" toml:"only,omitempty,inline" yaml:",omitempty"`
	Tags []string          `json:"tags,omitempty" mapstructure:"tags" toml:"tags,omitempty"        yaml:",omitempty"`
	Env  map[string]string `json:"env,omitempty"  mapstructure:"env"  toml:"env,omitempty"         yaml:",omitempty"`

	FileTypes []string `json:"file_types,omitempty" mapstructure:"file_types" toml:"file_types,omitempty" yaml:"file_types,omitempty"`

	Glob    string      `json:"glob,omitempty"    mapstructure:"glob"    toml:"glob,omitempty"    yaml:",omitempty"`
	Root    string      `json:"root,omitempty"    mapstructure:"root"    toml:"root,omitempty"    yaml:",omitempty"`
	Exclude interface{} `json:"exclude,omitempty" mapstructure:"exclude" toml:"exclude,omitempty" yaml:",omitempty"`

	Priority    int    `json:"priority,omitempty"    mapstructure:"priority"    toml:"priority,omitempty"    yaml:",omitempty"`
	FailText    string `json:"fail_text,omitempty"   mapstructure:"fail_text"   toml:"fail_text,omitempty"   yaml:"fail_text,omitempty"`
	Interactive bool   `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool   `json:"use_stdin,omitempty"   mapstructure:"use_stdin"   toml:"use_stdin,omitempty"   yaml:",omitempty"`
	StageFixed  bool   `json:"stage_fixed,omitempty" mapstructure:"stage_fixed" toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

type commandRunReplace struct {
	Run string `mapstructure:"run"`
}

func (c Command) Validate() error {
	if !isRunnerFilesCompatible(c.Run) {
		return errFilesIncompatible
	}

	return nil
}

func (c Command) DoSkip(gitState git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.check(gitState, c.Skip, c.Only)
}

func (c Command) ExecutionPriority() int {
	return c.Priority
}

func mergeCommands(base, extra *koanf.Koanf) (map[string]*Command, error) {
	if base == nil && extra == nil {
		return nil, nil
	}

	if base == nil {
		return unmarshalCommands(extra.Cut("commands"))
	}

	if extra == nil {
		return unmarshalCommands(base.Cut("commands"))
	}

	commandsOrigin := base.Cut("commands")
	commandsOverride := extra.Cut("commands")
	if commandsOrigin == nil {
		return unmarshalCommands(commandsOverride)
	}
	if commandsOverride == nil {
		return unmarshalCommands(commandsOrigin)
	}

	runReplaces := make(map[string]*commandRunReplace)
	for key := range commandsOrigin.Raw() {
		var replace commandRunReplace

		substructure := commandsOrigin.Cut(key)
		if substructure == nil {
			continue
		}

		if err := substructure.Unmarshal("", &replace); err != nil {
			return nil, err
		}

		runReplaces[key] = &replace
	}

	err := commandsOrigin.Merge(commandsOverride)
	if err != nil {
		return nil, err
	}

	commands, err := unmarshalCommands(commandsOrigin)
	if err != nil {
		return nil, err
	}

	for key, replace := range runReplaces {
		if replace.Run != "" {
			commands[key].Run = strings.ReplaceAll(commands[key].Run, CMD, replace.Run)
		}
	}

	return commands, nil
}

func unmarshalCommands(k *koanf.Koanf) (map[string]*Command, error) {
	commands := make(map[string]*Command)
	if err := k.Unmarshal("", &commands); err != nil {
		return nil, err
	}

	return commands, nil
}
