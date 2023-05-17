package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/evilmartians/lefthook/internal/version"
)

const dumpIndent = 2

type Config struct {
	Colors         interface{} `mapstructure:"colors" yaml:",omitempty"`
	Extends        []string    `mapstructure:"extends" yaml:",omitempty"`
	Remote         Remote      `mapstructure:"remote" yaml:",omitempty"`
	MinVersion     string      `mapstructure:"min_version" yaml:"min_version,omitempty"`
	SkipOutput     []string    `mapstructure:"skip_output" yaml:"skip_output,omitempty"`
	SourceDir      string      `mapstructure:"source_dir" yaml:"source_dir,omitempty"`
	SourceDirLocal string      `mapstructure:"source_dir_local" yaml:"source_dir_local,omitempty"`
	Rc             string      `mapstructure:"rc" yaml:",omitempty"`
	NoTTY          bool        `mapstructure:"no_tty" yaml:"no_tty,omitempty"`

	Hooks map[string]*Hook
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}

func (c *Config) Dump() error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(dumpIndent)
	defer encoder.Close()

	err := encoder.Encode(c)
	if err != nil {
		return err
	}

	return nil
}
