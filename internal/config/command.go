package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
)

var errFilesIncompatible = errors.New("One of your runners contains incompatible file types")

type Command struct {
	Run string `mapstructure:"run"`

	Skip  interface{}       `mapstructure:"skip"`
	Tags  []string          `mapstructure:"tags"`
	Glob  string            `mapstructure:"glob"`
	Files string            `mapstructure:"files"`
	Env   map[string]string `mapstructure:"env"`

	Root    string `mapstructure:"root"`
	Exclude string `mapstructure:"exclude"`

	FailText    string `mapstructure:"fail_text"`
	Interactive bool   `mapstructure:"interactive"`
}

func (c Command) Validate() error {
	if !isRunnerFilesCompatible(c.Run) {
		return errFilesIncompatible
	}

	return nil
}

func (c Command) DoSkip(gitState git.State) bool {
	if value := c.Skip; value != nil {
		return isSkip(gitState, value)
	}
	return false
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
