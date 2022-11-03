package config

type Remote struct {
	GitURL string `mapstructure:"git_url"`
	Ref    string `mapstructure:"ref"`
	Config string `mapstructure:"config"`
}

func (r Remote) Configured() bool {
	return len(r.GitURL) > 0
}
