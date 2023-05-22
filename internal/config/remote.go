package config

type Remote struct {
	GitURL string `mapstructure:"git_url" yaml:"git_url,omitempty" json:"git_url,omitempty" toml:"git_url"`
	Ref    string `mapstructure:"ref"     yaml:",omitempty"        json:"ref,omitempty" toml:"ref,omitempty"`
	Config string `mapstructure:"config"  yaml:",omitempty"        json:"config,omitempty" toml:"config,omitempty"`
}

func (r Remote) Configured() bool {
	return len(r.GitURL) > 0
}
