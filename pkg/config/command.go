package config

import (
	"github.com/spf13/viper"
)

type Command struct {
	Run string `mapstructure:"run"`

	Skip bool     `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`

	Root    string   `mapstructure:"root"`
	Exclude []string `mapstructure:"exclude"`

	// Deprecated
	Runner string `mapstructure:"runner"`
}

type commandRunReplace struct {
	Runner string `mapstructure:"runner"`
	Run    string `mapstructure:"run"`
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
	for key, _ := range commandsOrigin.AllSettings() {
		var replace commandRunReplace

		if err := commandsOrigin.Sub(key).Unmarshal(&replace); err != nil {
			return nil, err
		}

		runReplaces[key] = &replace
	}

	commandsOrigin.MergeConfigMap(commandsOverride.AllSettings())
	commands, err := unmarshalCommands(commandsOrigin)
	if err != nil {
		return nil, err
	}

	for key, replace := range runReplaces {
		// Deprecated, will be deleted
		if replace.Run != "" {
			commands[key].Run = replaceCmd(commands[key].Run, replace.Run)
		}

		if replace.Runner != "" {
			commands[key].Runner = replaceCmd(commands[key].Runner, replace.Runner)
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
