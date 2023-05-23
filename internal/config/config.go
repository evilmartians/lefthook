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

const dumpIndent = 2

type Config struct {
	MinVersion     string      `mapstructure:"min_version,omitempty"`
	SourceDir      string      `mapstructure:"source_dir"`
	SourceDirLocal string      `mapstructure:"source_dir_local"`
	Rc             string      `mapstructure:"rc,omitempty"`
	SkipOutput     []string    `mapstructure:"skip_output,omitempty"`
	Extends        []string    `mapstructure:"extends,omitempty"`
	NoTTY          bool        `mapstructure:"no_tty,omitempty"`
	Colors         interface{} `mapstructure:"colors,omitempty"`
	Remote         *Remote     `mapstructure:"remote,omitempty" `

	Hooks map[string]*Hook `mapstructure:"-"`
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}

func (c *Config) Dump(asJSON bool, asTOML bool) error {
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

	if asJSON {
		return dumpJSON(res)
	}

	if asTOML {
		return dumpTOML(res)
	}

	return dumpYAML(res)
}

func dumpYAML(input map[string]interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(dumpIndent)
	defer encoder.Close()

	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}

func dumpJSON(input map[string]interface{}) error {
	res, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}

	log.Info(string(res))

	return nil
}

func dumpTOML(input map[string]interface{}) error {
	encoder := toml.NewEncoder(os.Stdout)
	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}
