package config

type Remote struct {
	GitURL string `mapstructure:"git_url"           yaml:"git_url"    json:"git_url,omitempty" toml:"git_url"`
	Ref    string `mapstructure:"ref,omitempty"     yaml:",omitempty" json:"ref,omitempty"     toml:"ref,omitempty"`
	Config string `mapstructure:"config,omitempty"  yaml:",omitempty" json:"config,omitempty"  toml:"config,omitempty"`
}

func (r *Remote) Configured() bool {
	if r == nil {
		return false
	}

	return len(r.GitURL) > 0
}
