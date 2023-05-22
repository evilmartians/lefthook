package config

import (
	"encoding/json"
	"os"

	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const dumpIndent = 2

type Config struct {
	Colors         interface{} `mapstructure:"colors"           yaml:"colors,omitempty"           toml:"colors,omitempty"           json:"colors,omitempty"`
	Extends        []string    `mapstructure:"extends"          yaml:"extends,omitempty"          toml:"extends,omitempty"          json:"extends,omitempty"`
	Remote         Remote      `mapstructure:"remote"           yaml:"remote,omitempty"           toml:"remote,omitempty"           json:"remote,omitempty"`
	MinVersion     string      `mapstructure:"min_version"      yaml:"min_version,omitempty"      toml:"min_version,omitempty"      json:"min_version,omitempty"`
	SkipOutput     []string    `mapstructure:"skip_output"      yaml:"skip_output,omitempty"      toml:"skip_output,omitempty"      json:"skip_output,omitempty"`
	SourceDir      string      `mapstructure:"source_dir"       yaml:"source_dir,omitempty"       toml:"source_dir,omitempty"       json:"source_dir,omitempty"`
	SourceDirLocal string      `mapstructure:"source_dir_local" yaml:"source_dir_local,omitempty" toml:"source_dir_local,omitempty" json:"source_dir_local,omitempty"`
	Rc             string      `mapstructure:"rc"               yaml:"rc,omitempty"               toml:"rc,omitempty"               json:"rc,omitempty"`
	NoTTY          bool        `mapstructure:"no_tty"           yaml:"no_tty,omitempty"           toml:"no_tty,omitempty"           json:"no_tty,omitempty"`

	Hooks map[string]*Hook `yaml:",inline" json:"-"`
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}

func (c *Config) Dump(asJSON bool, asTOML bool) error {
	if asJSON {
		return c.dumpJSON()
	}

	if asTOML {
		return c.dumpTOML()
	}

	return c.dumpYAML()
}

func (c *Config) dumpYAML() error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(dumpIndent)
	defer encoder.Close()

	err := encoder.Encode(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) dumpJSON() error {
	// This hack allows to inline Hooks
	type ConfigForMarshalling *Config
	res, err := json.Marshal(ConfigForMarshalling(c))
	if err != nil {
		return err
	}

	var rawMarshalled map[string]json.RawMessage
	if err = json.Unmarshal(res, &rawMarshalled); err != nil {
		return err
	}

	for hook, contents := range c.Hooks {
		var hookMarshalled json.RawMessage
		hookMarshalled, err = json.Marshal(contents)
		if err != nil {
			return err
		}
		rawMarshalled[hook] = hookMarshalled
	}

	res, err = json.MarshalIndent(rawMarshalled, "", "  ")
	if err != nil {
		return err
	}

	log.Info(string(res))

	return nil
}

func (c *Config) dumpTOML() error {
	// type ConfigForMarshalling *Config
	// res, err := toml.Marshal(ConfigForMarshalling(c))
	// if err != nil {
	// 	return err
	// }

	// var rawMarshalled map[string]interface{}
	// if err = json.Unmarshal(res, &rawMarshalled); err != nil {
	// 	return err
	// }

	// for hook, contents := range c.Hooks {
	// 	var hookMarshalled interface{}
	// 	hookMarshalled, err = toml.Marshal(contents)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	rawMarshalled[hook] = hookMarshalled
	// }

	// res, err = toml.Marshal(rawMarshalled)
	// if err != nil {
	// 	return err
	// }

	// log.Info(string(res))

	// return nil

	res := make(map[string]interface{})

	if c.Colors != nil {
		res["colors"] = c.Colors
	}
	if len(c.Extends) != 0 {
		res["extends"] = c.Extends
	}
	if len(c.Remote.GitURL) != 0 {
		res["remote"] = c.Remote
	}
	if len(c.MinVersion) != 0 {
		res["min_version"] = c.MinVersion
	}
	if len(c.SkipOutput) != 0 {
		res["skip_output"] = c.SkipOutput
	}
	if c.SourceDir != DefaultSourceDir {
		res["source_dir"] = c.SourceDir
	}
	if c.SourceDirLocal != DefaultSourceDirLocal {
		res["source_dir_local"] = c.SourceDirLocal
	}
	if len(c.Rc) != 0 {
		res["rc"] = c.Rc
	}
	if c.NoTTY {
		res["no_tty"] = c.NoTTY
	}

	for hookName, hook := range c.Hooks {
		res[hookName] = hook
	}

	encoder := toml.NewEncoder(os.Stdout)
	// err := encoder.Encode(c)
	err := encoder.Encode(res)
	if err != nil {
		return err
	}

	return nil
}
