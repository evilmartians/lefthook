package config

import (
	"errors"

	"github.com/evilmartians/lefthook/internal/version"
)

var errInvalidVersion = errors.New("Current Lefhook version is lower than config version or 'min_version' is incorrect. Check the format: '0.9.9'")

type Config struct {
	Colors         bool     `mapstructure:"colors"`
	Extends        []string `mapstructure:"extends"`
	MinVersion     string   `mapstructure:"min_version"`
	SkipOutput     []string `mapstructure:"skip_output"`
	SourceDir      string   `mapstructure:"source_dir"`
	SourceDirLocal string   `mapstructure:"source_dir_local"`

	Hooks map[string]*Hook
}

func (c *Config) Validate() error {
	if !version.IsCovered(c.MinVersion) {
		return errInvalidVersion
	}

	return nil
}
