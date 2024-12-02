package config

import (
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

func (s Script) DoSkip(state func() git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.check(state, s.Skip, s.Only)
}

func (s Script) ExecutionPriority() int {
	return s.Priority
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
// func unmarshal(input, output interface{}) error {
// 	return mapstructure.WeakDecode(input, &output)
// }
