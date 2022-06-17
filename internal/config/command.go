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

	Skip  interface{} `mapstructure:"skip"`
	Tags  []string    `mapstructure:"tags"`
	Glob  string      `mapstructure:"glob"`
	Files string      `mapstructure:"files"`

	Root    string `mapstructure:"root"`
	Exclude string `mapstructure:"exclude"`

	// DEPRECATED
	Runner string `mapstructure:"runner"`
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
	Run    string `mapstructure:"run"`
	Runner string `mapstructure:"runner"` // DEPRECATED
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

		if err := commandsOrigin.Sub(key).Unmarshal(&replace); err != nil {
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
		// Deprecated, will be deleted
		if replace.Run != "" {
			commands[key].Run = strings.Replace(commands[key].Run, CMD, replace.Run, -1)
		}

		if replace.Runner != "" {
			commands[key].Runner = strings.Replace(commands[key].Runner, CMD, replace.Runner, -1)
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
