package config

import (
	"github.com/evilmartians/lefthook/internal/version"
)

type Config struct {
	Colors         bool     `mapstructure:"colors"`
	Extends        []string `mapstructure:"extends"`
	Remotes        []Remote `mapstructure:"remotes"`
	MinVersion     string   `mapstructure:"min_version"`
	SkipOutput     []string `mapstructure:"skip_output"`
	SourceDir      string   `mapstructure:"source_dir"`
	SourceDirLocal string   `mapstructure:"source_dir_local"`

	Hooks map[string]*Hook
}

type Remote struct {
	URL     string   `mapstructure:"url"`
	Rev     string   `mapstructure:"rev"`
	Configs []string `mapstructure:"configs"`
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}
