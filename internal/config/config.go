package config

import (
	"github.com/evilmartians/lefthook/internal/version"
)

type Config struct {
	Colors         bool     `mapstructure:"colors"`
	Extends        []string `mapstructure:"extends"`
	Remote         Remote   `mapstructure:"remote"`
	MinVersion     string   `mapstructure:"min_version"`
	SkipOutput     []string `mapstructure:"skip_output"`
	SourceDir      string   `mapstructure:"source_dir"`
	SourceDirLocal string   `mapstructure:"source_dir_local"`
	Rc             string   `mapstructure:"rc"`
	NoTTY          bool     `mapstructure:"no_tty"`

	Hooks map[string]*Hook
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}
