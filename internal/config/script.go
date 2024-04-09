package config

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

type Script struct {
	Runner string `json:"runner" mapstructure:"runner" toml:"runner" yaml:"runner"`

	Skip     interface{}       `json:"skip,omitempty"     mapstructure:"skip"     toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only     interface{}       `json:"only,omitempty"     mapstructure:"only"     toml:"only,omitempty,inline" yaml:",omitempty"`
	Tags     []string          `json:"tags,omitempty"     mapstructure:"tags"     toml:"tags,omitempty"        yaml:",omitempty"`
	Env      map[string]string `json:"env,omitempty"      mapstructure:"env"      toml:"env,omitempty"         yaml:",omitempty"`
	Priority int               `json:"priority,omitempty" mapstructure:"priority" toml:"priority,omitempty"    yaml:",omitempty"`

	FailText    string `json:"fail_text,omitempty"   mapstructure:"fail_text"   toml:"fail_text,omitempty"   yaml:"fail_text,omitempty"`
	Interactive bool   `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool   `json:"use_stdin,omitempty"   mapstructure:"use_stdin"   toml:"use_stdin,omitempty"   yaml:",omitempty"`
	StageFixed  bool   `json:"stage_fixed,omitempty" mapstructure:"stage_fixed" toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

type scriptRunnerReplace struct {
	Runner string `mapstructure:"runner"`
}

func (s Script) DoSkip(gitState git.State) bool {
	skipChecker := NewSkipChecker(system.Executor{})
	return skipChecker.check(gitState, s.Skip, s.Only)
}

func (s Script) ExecutionPriority() int {
	return s.Priority
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
