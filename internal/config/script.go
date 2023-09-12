package config

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
)

type Script struct {
	Runner string `mapstructure:"runner" yaml:"runner" json:"runner" toml:"runner"`

	Skip interface{}       `mapstructure:"skip" yaml:",omitempty" json:"skip,omitempty" toml:"skip,omitempty,inline"`
	Only interface{}       `mapstructure:"only" yaml:",omitempty" json:"only,omitempty" toml:"only,omitempty,inline"`
	Tags []string          `mapstructure:"tags" yaml:",omitempty" json:"tags,omitempty" toml:"tags,omitempty"`
	Env  map[string]string `mapstructure:"env"  yaml:",omitempty" json:"env,omitempty"  toml:"env,omitempty"`

	FailText    string `mapstructure:"fail_text"   yaml:"fail_text,omitempty"   json:"fail_text,omitempty"   toml:"fail_text,omitempty"`
	Interactive bool   `mapstructure:"interactive" yaml:",omitempty"            json:"interactive,omitempty" toml:"interactive,omitempty"`
	UseStdin    bool   `mapstructure:"use_stdin"   yaml:",omitempty"            json:"use_stdin,omitempty"   toml:"use_stdin,omitempty"`
	StageFixed  bool   `mapstructure:"stage_fixed" yaml:"stage_fixed,omitempty" json:"stage_fixed,omitempty" toml:"stage_fixed,omitempty"`
}

func (s Script) DoSkip(gitState git.State) bool {
	return doSkip(gitState, s.Skip, s.Only)
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
	return mapstructure.WeakDecode(input, &output)
}
