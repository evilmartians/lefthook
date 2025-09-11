package config

import (
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

const CMD = "{cmd}"

type Hook struct {
	Parallel      bool        `json:"parallel,omitempty"        mapstructure:"parallel"                                                        toml:"parallel,omitempty"   yaml:",omitempty"`
	Piped         bool        `json:"piped,omitempty"           mapstructure:"piped"                                                           toml:"piped,omitempty"      yaml:",omitempty"`
	Follow        bool        `json:"follow,omitempty"          mapstructure:"follow"                                                          toml:"follow,omitempty"     yaml:",omitempty"`
	FailOnChanges string      `json:"fail_on_changes,omitempty" jsonschema:"enum=true,enum=1,enum=0,enum=false,enum=never,enum=always,enum=ci" koanf:"fail_on_changes"     mapstructure:"fail_on_changes" toml:"fail_on_changes,omitempty" yaml:"fail_on_changes,omitempty"`
	Files         string      `json:"files,omitempty"           mapstructure:"files"                                                           toml:"files,omitempty"      yaml:",omitempty"`
	ExcludeTags   []string    `json:"exclude_tags,omitempty"    koanf:"exclude_tags"                                                           mapstructure:"exclude_tags" toml:"exclude_tags,omitempty"  yaml:"exclude_tags,omitempty"`
	Skip          interface{} `json:"skip,omitempty"            jsonschema:"oneof_type=boolean;array"                                          mapstructure:"skip"         toml:"skip,omitempty,inline"   yaml:",omitempty"`
	Only          interface{} `json:"only,omitempty"            jsonschema:"oneof_type=boolean;array"                                          mapstructure:"only"         toml:"only,omitempty,inline"   yaml:",omitempty"`

	Jobs []*Job `json:"jobs,omitempty" mapstructure:"jobs" toml:"jobs,omitempty" yaml:",omitempty"`

	Commands map[string]*Command `json:"commands,omitempty" mapstructure:"-" toml:"commands,omitempty" yaml:",omitempty"`
	Scripts  map[string]*Script  `json:"scripts,omitempty"  mapstructure:"-" toml:"scripts,omitempty"  yaml:",omitempty"`
}

func (h *Hook) DoSkip(state func() git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.Check(state, h.Skip, h.Only)
}
