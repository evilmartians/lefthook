package config

import (
	"errors"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

var errFilesIncompatible = errors.New("One of your runners contains incompatible file types")

type Command struct {
	Run   string `json:"run"             mapstructure:"run"   toml:"run"             yaml:"run"`
	Files string `json:"files,omitempty" mapstructure:"files" toml:"files,omitempty" yaml:",omitempty"`

	Skip interface{}       `json:"skip,omitempty" mapstructure:"skip" toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only interface{}       `json:"only,omitempty" mapstructure:"only" toml:"only,omitempty,inline" yaml:",omitempty"`
	Tags []string          `json:"tags,omitempty" mapstructure:"tags" toml:"tags,omitempty"        yaml:",omitempty"`
	Env  map[string]string `json:"env,omitempty"  mapstructure:"env"  toml:"env,omitempty"         yaml:",omitempty"`

	FileTypes []string `json:"file_types,omitempty" koanf:"file_types" mapstructure:"file_types" toml:"file_types,omitempty" yaml:"file_types,omitempty"`

	Glob    string      `json:"glob,omitempty"    mapstructure:"glob"    toml:"glob,omitempty"    yaml:",omitempty"`
	Root    string      `json:"root,omitempty"    mapstructure:"root"    toml:"root,omitempty"    yaml:",omitempty"`
	Exclude interface{} `json:"exclude,omitempty" mapstructure:"exclude" toml:"exclude,omitempty" yaml:",omitempty"`

	Priority    int    `json:"priority,omitempty"    mapstructure:"priority"    toml:"priority,omitempty"    yaml:",omitempty"`
	FailText    string `json:"fail_text,omitempty"   koanf:"fail_text"          mapstructure:"fail_text"     toml:"fail_text,omitempty"   yaml:"fail_text,omitempty"`
	Interactive bool   `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool   `json:"use_stdin,omitempty"   mapstructure:"use_stdin"   toml:"use_stdin,omitempty"   yaml:",omitempty"`
	StageFixed  bool   `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

func (c Command) Validate() error {
	if !isRunnerFilesCompatible(c.Run) {
		return errFilesIncompatible
	}

	return nil
}

func (c Command) DoSkip(state func() git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.check(state, c.Skip, c.Only)
}

func (c Command) ExecutionPriority() int {
	return c.Priority
}
