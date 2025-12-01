package config

type Job struct {
	Name     string `json:"name,omitempty"      mapstructure:"name"                       toml:"name,omitempty"    yaml:",omitempty"`
	Run      string `json:"run,omitempty"       jsonschema:"oneof_required=Run a command" mapstructure:"run"       toml:"run,omitempty"       yaml:",omitempty"`
	Script   string `json:"script,omitempty"    jsonschema:"oneof_required=Run a script"  mapstructure:"script"    toml:"script,omitempty"    yaml:",omitempty"`
	Runner   string `json:"runner,omitempty"    mapstructure:"runner"                     toml:"runner,omitempty"  yaml:",omitempty"`
	Args     string `json:"args,omitempty"      mapstructure:"args"                       toml:"args,omitempty"    yaml:",omitempty"`
	Root     string `json:"root,omitempty"      mapstructure:"root"                       toml:"root,omitempty"    yaml:",omitempty"`
	Files    string `json:"files,omitempty"     mapstructure:"files"                      toml:"files,omitempty"   yaml:",omitempty"`
	FailText string `json:"fail_text,omitempty" koanf:"fail_text"                         mapstructure:"fail_text" toml:"fail_text,omitempty" yaml:"fail_text,omitempty"`

	Glob      []string `json:"glob,omitempty"       jsonschema:"oneof_type=string;array" mapstructure:"glob"       toml:"glob,omitempty"       yaml:",omitempty"`
	Exclude   []string `json:"exclude,omitempty"    jsonschema:"oneof_type=string;array" mapstructure:"exclude"    toml:"exclude,omitempty"    yaml:",omitempty"`
	Tags      []string `json:"tags,omitempty"       mapstructure:"tags"                  toml:"tags,omitempty"     yaml:",omitempty"`
	FileTypes []string `json:"file_types,omitempty" koanf:"file_types"                   mapstructure:"file_types" toml:"file_types,omitempty" yaml:"file_types,omitempty"`

	Env map[string]string `json:"env,omitempty" mapstructure:"env" toml:"env,omitempty" yaml:",omitempty"`

	Interactive bool `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool `json:"use_stdin,omitempty"   koanf:"use_stdin"          mapstructure:"use_stdin"     toml:"use_stdin,omitempty"   yaml:"use_stdin,omitempty"`
	StageFixed  bool `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`

	Skip any `json:"skip,omitempty" jsonschema:"oneof_type=boolean;array" mapstructure:"skip" toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only any `json:"only,omitempty" jsonschema:"oneof_type=boolean;array" mapstructure:"only" toml:"only,omitempty,inline" yaml:",omitempty"`

	Group *Group `json:"group,omitempty" jsonschema:"oneof_required=Run a group" mapstructure:"group" toml:"group,omitempty" yaml:",omitempty"`
}

type Group struct {
	Root     string `json:"root,omitempty"     mapstructure:"root"     toml:"root,omitempty"     yaml:",omitempty"`
	Parallel bool   `json:"parallel,omitempty" mapstructure:"parallel" toml:"parallel,omitempty" yaml:",omitempty"`
	Piped    bool   `json:"piped,omitempty"    mapstructure:"piped"    toml:"piped,omitempty"    yaml:",omitempty"`
	Jobs     []*Job `json:"jobs"               mapstructure:"jobs"     toml:"jobs"               yaml:"jobs"`
}

func (job *Job) PrintableName(id string) string {
	if len(job.Name) != 0 {
		return job.Name
	}
	if len(job.Run) != 0 {
		return job.Run
	}
	if len(job.Script) != 0 {
		return job.Script
	}

	return "[" + id + "]"
}
