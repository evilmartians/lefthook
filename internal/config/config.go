package config

import (
	"encoding/json"
	"os"

	"github.com/mitchellh/mapstructure"
	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

type DumpFormat int

const (
	YAMLFormat DumpFormat = iota
	JSONFormat
	TOMLFormat

	dumpIndent = 2
)

type Config struct {
	MinVersion              string      `mapstructure:"min_version,omitempty"`
	SourceDir               string      `mapstructure:"source_dir"`
	SourceDirLocal          string      `mapstructure:"source_dir_local"`
	Rc                      string      `mapstructure:"rc,omitempty"`
	SkipOutput              interface{} `mapstructure:"skip_output,omitempty"`
	Output                  interface{} `mapstructure:"output,omitempty"`
	Extends                 []string    `mapstructure:"extends,omitempty"`
	NoTTY                   bool        `mapstructure:"no_tty,omitempty"`
	AssertLefthookInstalled bool        `mapstructure:"assert_lefthook_installed,omitempty"`
	Colors                  interface{} `mapstructure:"colors,omitempty"`

	// Deprecated: use Remotes
	Remote  *Remote   `mapstructure:"remote,omitempty"`
	Remotes []*Remote `mapstructure:"remotes,omitempty"`

	Hooks map[string]*Hook `mapstructure:"-"`
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}

func (c *Config) Dump(format DumpFormat) error {
	res := make(map[string]interface{})
	if err := mapstructure.Decode(c, &res); err != nil {
		return err
	}

	if c.SourceDir == DefaultSourceDir {
		delete(res, "source_dir")
	}
	if c.SourceDirLocal == DefaultSourceDirLocal {
		delete(res, "source_dir_local")
	}

	for hookName, hook := range c.Hooks {
		res[hookName] = hook
	}

	var dumper dumper
	switch format {
	case YAMLFormat:
		dumper = yamlDumper{}
	case TOMLFormat:
		dumper = tomlDumper{}
	case JSONFormat:
		dumper = jsonDumper{}
	default:
		dumper = yamlDumper{}
	}

	return dumper.Dump(res)
}

type dumper interface {
	Dump(map[string]interface{}) error
}

type yamlDumper struct{}

func (yamlDumper) Dump(input map[string]interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(dumpIndent)
	defer encoder.Close()

	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}

type tomlDumper struct{}

func (tomlDumper) Dump(input map[string]interface{}) error {
	encoder := toml.NewEncoder(os.Stdout)
	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}

type jsonDumper struct{}

func (jsonDumper) Dump(input map[string]interface{}) error {
	res, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}

	log.Info(string(res))

	return nil
}
