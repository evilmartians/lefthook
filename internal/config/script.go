package config

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
)

type Script struct {
	Runner string `mapstructure:"runner"`

	Skip interface{}       `mapstructure:"skip"`
	Tags []string          `mapstructure:"tags"`
	Env  map[string]string `mapstructure:"env"`

	FailText    string `mapstructure:"fail_text"`
	Interactive bool   `mapstructure:"interactive"`
}

func (s Script) DoSkip(gitState git.State) bool {
	if value := s.Skip; value != nil {
		return isSkip(gitState, value)
	}
	return false
}

type scriptRunnerReplace struct {
	Runner string `mapstructure:"runner"`
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

	err := base.MergeConfigMap(map[string]interface{}{
		"scripts": scriptsOverride,
	})
	if err != nil {
		return nil, err
	}

	scripts, err := unmarshalScripts(base.GetStringMap("scripts"))
	if err != nil {
		return nil, err
	}

	for key, replace := range runReplaces {
		if replace.Runner != "" {
			scripts[key].Runner = strings.ReplaceAll(scripts[key].Runner, CMD, replace.Runner)
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
// ```yaml
// scripts:
//
//	"example.sh":
//	    runner: bash
//
// ```
//
// Unmarshals into this:
//
// ```yaml
// scripts:
//
//	example:
//	  sh:
//	    runner: bash
//
// ```
//
// This is not an expected behavior and cannot be controlled yet
// Working with GetStringMap is the only way to get the structure "as is".
func unmarshal(input, output interface{}) error {
	if err := mapstructure.WeakDecode(input, &output); err != nil {
		return err
	}

	return nil
}
