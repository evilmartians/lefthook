package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Script struct {
	Runner string `mapstructure:"runner"`

	Skip bool     `mapstructure:"skip"`
	Tags []string `mapstructure:"tags"`

	// Deprecated
	Run string `mapstructure:"run"`
}

type scriptRunnerReplace struct {
	Runner string `mapstructure:"runner"`
	Run    string `mapstructure:"run"` // Deprecated
}

func mergeScripts(base, extra *viper.Viper) (map[string]*Script, error) {
	if base == nil && extra == nil {
		return nil, nil
	}

	if base == nil {
		return unmarshalScripts(extra.GetStringMap("scripts"))
	}

	if extra == nil {
		return unmarshalScripts(base.GetStringMap("scripts"))
	}

	scriptsOrigin := base.GetStringMap("scripts")
	scriptsOverride := extra.GetStringMap("scripts")
	if scriptsOrigin == nil {
		return unmarshalScripts(scriptsOverride)
	}
	if scriptsOverride == nil {
		return unmarshalScripts(scriptsOrigin)
	}

	runReplaces := make(map[string]*scriptRunnerReplace)
	for key, originConfig := range scriptsOrigin {
		var runReplace scriptRunnerReplace

		if err := unmarshal(originConfig, &runReplace); err != nil {
			return nil, err
		}

		runReplaces[key] = &runReplace
	}

	base.MergeConfigMap(map[string]interface{}{
		"scripts": scriptsOverride,
	})

	scripts, err := unmarshalScripts(base.GetStringMap("scripts"))
	if err != nil {
		return nil, err
	}

	for key, replace := range runReplaces {
		// Deprecated, will be deleted
		if replace.Run != "" {
			scripts[key].Run = replaceCmd(scripts[key].Run, replace.Run)
		}

		if replace.Runner != "" {
			scripts[key].Runner = replaceCmd(scripts[key].Runner, replace.Runner)
		}
	}

	return scripts, nil
}

func unmarshalScripts(s map[string]interface{}) (map[string]*Script, error) {
	if len(s) == 0 {
		return nil, nil
	}

	scripts := make(map[string]*Script)
	for name, scriptConfig := range s {
		var script Script

		if err := unmarshal(scriptConfig, &script); err != nil {
			return nil, err
		}

		scripts[name] = &script
	}

	return scripts, nil
}

// `scripts` are unmarshalled manually because viper
// uses "." as a key delimiter. So, this definition:
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
// Working with GetStringMap is the only way to get the structure "as is"
func unmarshal(input, output interface{}) error {
	if err := mapstructure.WeakDecode(input, &output); err != nil {
		return err
	}

	return nil
}
