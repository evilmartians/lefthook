package config

type Job struct {
	Name   string `json:"name,omitempty"   mapstructure:"name"   toml:"name,omitempty"   yaml:",omitempty"`
	Run    string `json:"run,omitempty"    mapstructure:"run"    toml:"run,omitempty"    yaml:",omitempty"`
	Script string `json:"script,omitempty" mapstructure:"script" toml:"script,omitempty" yaml:",omitempty"`
	Runner string `json:"runner,omitempty" mapstructure:"runner" toml:"runner,omitempty" yaml:",omitempty"`

	Glob     string `json:"glob,omitempty"      mapstructure:"glob"  toml:"glob,omitempty"    yaml:",omitempty"`
	Root     string `json:"root,omitempty"      mapstructure:"root"  toml:"root,omitempty"    yaml:",omitempty"`
	Files    string `json:"files,omitempty"     mapstructure:"files" toml:"files,omitempty"   yaml:",omitempty"`
	FailText string `json:"fail_text,omitempty" koanf:"fail_text"    mapstructure:"fail_text" toml:"fail_text,omitempty" yaml:"fail_text,omitempty"`

	Tags      []string `json:"tags,omitempty"       mapstructure:"tags" toml:"tags,omitempty"     yaml:",omitempty"`
	FileTypes []string `json:"file_types,omitempty" koanf:"file_types"  mapstructure:"file_types" toml:"file_types,omitempty" yaml:"file_types,omitempty"`

	Env map[string]string `json:"env,omitempty" mapstructure:"env" toml:"env,omitempty" yaml:",omitempty"`

	Interactive bool `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool `json:"use_stdin,omitempty"   koanf:"use_stdin"          mapstructure:"use_stdin"     toml:"use_stdin,omitempty"   yaml:",omitempty"`
	StageFixed  bool `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`

	Exclude interface{} `json:"exclude,omitempty" mapstructure:"exclude" toml:"exclude,omitempty"     yaml:",omitempty"`
	Skip    interface{} `json:"skip,omitempty"    mapstructure:"skip"    toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only    interface{} `json:"only,omitempty"    mapstructure:"only"    toml:"only,omitempty,inline" yaml:",omitempty"`

	Group *Group `json:"group,omitempty" mapstructure:"group" toml:"group,omitempty" yaml:",omitempty"`
}

type Group struct {
	Root     string `json:"root,omitempty"     mapstructure:"root"     toml:"root,omitempty"     yaml:",omitempty"`
	Parallel bool   `json:"parallel,omitempty" mapstructure:"parallel" toml:"parallel,omitempty" yaml:",omitempty"`
	Piped    bool   `json:"piped,omitempty"    mapstructure:"piped"    toml:"piped,omitempty"    yaml:",omitempty"`
	Jobs     []*Job `json:"jobs,omitempty"     mapstructure:"jobs"     toml:"jobs,omitempty"     yaml:",omitempty"`
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
