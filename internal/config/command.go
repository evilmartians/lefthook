package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
)

var errFilesIncompatible = errors.New("One of your runners contains incompatible file types")

type Command struct {
	Run string `mapstructure:"run" yaml:"run" json:"run" toml:"run"`

	Skip  interface{}       `mapstructure:"skip"  yaml:",omitempty"  json:"skip,omitempty"  toml:"skip,omitempty,inline"`
	Only  interface{}       `mapstructure:"only"  yaml:",omitempty"  json:"only,omitempty"  toml:"only,omitempty,inline"`
	Tags  []string          `mapstructure:"tags"  yaml:",omitempty"  json:"tags,omitempty"  toml:"tags,omitempty"`
	Glob  string            `mapstructure:"glob"  yaml:",omitempty"  json:"glob,omitempty"  toml:"glob,omitempty"`
	Files string            `mapstructure:"files" yaml:",omitempty"  json:"files,omitempty" toml:"files,omitempty"`
	Env   map[string]string `mapstructure:"env"   yaml:",omitempty"  json:"env,omitempty"   toml:"env,omitempty"`

	Root    string `mapstructure:"root"    yaml:",omitempty" json:"root,omitempty"    toml:"root,omitempty"`
	Exclude string `mapstructure:"exclude" yaml:",omitempty" json:"exclude,omitempty" toml:"exclude,omitempty"`

	FailText    string `mapstructure:"fail_text"   yaml:"fail_text,omitempty"   json:"fail_text,omitempty"   toml:"fail_text,omitempty"`
	Interactive bool   `mapstructure:"interactive" yaml:",omitempty"            json:"interactive,omitempty" toml:"interactive,omitempty"`
	UseStdin    bool   `mapstructure:"use_stdin"   yaml:",omitempty"            json:"use_stdin,omitempty"   toml:"use_stdin,omitempty"`
	StageFixed  bool   `mapstructure:"stage_fixed" yaml:"stage_fixed,omitempty" json:"stage_fixed,omitempty" toml:"stage_fixed,omitempty"`
}

func (c Command) Validate() error {
	if !isRunnerFilesCompatible(c.Run) {
		return errFilesIncompatible
	}

	return nil
}

func (c Command) DoSkip(gitState git.State) bool {
	return doSkip(gitState, c.Skip, c.Only)
}

type commandRunReplace struct {
	Run string `mapstructure:"run"`
}

func mergeCommands(base, extra *viper.Viper) (map[string]*Command, error) {
	if base == nil && extra == nil {
		return nil, nil
	}

	if base == nil {
		return unmarshalCommands(extra.Sub("commands"))
	}

	if extra == nil {
		return unmarshalCommands(base.Sub("commands"))
	}

	commandsOrigin := base.Sub("commands")
	commandsOverride := extra.Sub("commands")
	if commandsOrigin == nil {
		return unmarshalCommands(commandsOverride)
	}
	if commandsOverride == nil {
		return unmarshalCommands(commandsOrigin)
	}

	runReplaces := make(map[string]*commandRunReplace)
	for key := range commandsOrigin.AllSettings() {
		var replace commandRunReplace

		substructure := commandsOrigin.Sub(key)
		if substructure == nil {
			continue
		}

		if err := substructure.Unmarshal(&replace); err != nil {
			return nil, err
		}

		runReplaces[key] = &replace
	}

	err := commandsOrigin.MergeConfigMap(commandsOverride.AllSettings())
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

func unmarshalCommands(v *viper.Viper) (map[string]*Command, error) {
	if v == nil {
		return nil, nil
	}

	commands := make(map[string]*Command)
	if err := v.Unmarshal(&commands); err != nil {
		return nil, err
	}

	return commands, nil
}
