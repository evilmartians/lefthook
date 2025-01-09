package config

import (
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

type Script struct {
	Runner string `json:"runner" mapstructure:"runner" toml:"runner" yaml:"runner"`

	Skip     interface{}       `json:"skip,omitempty"     jsonschema:"oneof_type=boolean;array" mapstructure:"skip"       toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only     interface{}       `json:"only,omitempty"     jsonschema:"oneof_type=boolean;array" mapstructure:"only"       toml:"only,omitempty,inline" yaml:",omitempty"`
	Tags     []string          `json:"tags,omitempty"     mapstructure:"tags"                   toml:"tags,omitempty"     yaml:",omitempty"`
	Env      map[string]string `json:"env,omitempty"      mapstructure:"env"                    toml:"env,omitempty"      yaml:",omitempty"`
	Priority int               `json:"priority,omitempty" mapstructure:"priority"               toml:"priority,omitempty" yaml:",omitempty"`

	FailText    string `json:"fail_text,omitempty"   koanf:"fail_text"          mapstructure:"fail_text"     toml:"fail_text,omitempty"   yaml:"fail_text,omitempty"`
	Interactive bool   `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool   `json:"use_stdin,omitempty"   koanf:"use_stdin"          mapstructure:"use_stdin"     toml:"use_stdin,omitempty"   yaml:"use_stdin,omitempty"`
	StageFixed  bool   `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

func (s Script) DoSkip(state func() git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.Check(state, s.Skip, s.Only)
}

func (s Script) ExecutionPriority() int {
	return s.Priority
}
