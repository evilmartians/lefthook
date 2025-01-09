package config

type Remote struct {
	// A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on.
	GitURL string `json:"git_url,omitempty" koanf:"git_url" mapstructure:"git_url" toml:"git_url" yaml:"git_url"`

	// An optional *branch* or *tag* name
	Ref string `json:"ref,omitempty" mapstructure:"ref,omitempty" toml:"ref,omitempty" yaml:",omitempty"`

	// An optional array of config paths from remote's root
	Configs []string `json:"configs,omitempty" mapstructure:"configs,omitempty" toml:"configs,omitempty" yaml:",omitempty"`

	// Set to true if you want to always refetch the remote
	Refetch bool `json:"refetch,omitempty" mapstructure:"refetch,omitempty" toml:"refetch,omitempty" yaml:",omitempty"`

	// Provide a frequency for the remotes refetches
	RefetchFrequency string `json:"refetch_frequency,omitempty" jsonschema:"example=24h" koanf:"refetch_frequency" mapstructure:"refetch_frequency,omitempty" toml:"refetch_frequency,omitempty" yaml:",omitempty"`

	// Deprecated: use `configs`
	Config string `json:"config,omitempty" mapstructure:"config,omitempty" toml:"config,omitempty" yaml:",omitempty"`
}

func (r *Remote) Configured() bool {
	if r == nil {
		return false
	}

	return len(r.GitURL) > 0
}
