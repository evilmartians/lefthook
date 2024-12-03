package config

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"
	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/evilmartians/lefthook/internal/version"
)

type DumpFormat int

const (
	YAMLFormat DumpFormat = iota
	TOMLFormat
	JSONFormat
	JSONCompactFormat

	yamlIndent = 2
)

type Config struct {
	MinVersion              string      `koanf:"min_version"               mapstructure:"min_version,omitempty"`
	SourceDir               string      `koanf:"source_dir"                mapstructure:"source_dir"`
	SourceDirLocal          string      `koanf:"source_dir_local"          mapstructure:"source_dir_local"`
	Rc                      string      `mapstructure:"rc,omitempty"`
	SkipOutput              interface{} `koanf:"skip_output"               mapstructure:"skip_output,omitempty"`
	Output                  interface{} `mapstructure:"output,omitempty"`
	Extends                 []string    `mapstructure:"extends,omitempty"`
	NoTTY                   bool        `koanf:"no_tty"                    mapstructure:"no_tty,omitempty"`
	AssertLefthookInstalled bool        `koanf:"assert_lefthook_installed" mapstructure:"assert_lefthook_installed,omitempty"`
	Colors                  interface{} `mapstructure:"colors,omitempty"`
	SkipLFS                 bool        `mapstructure:"skip_lfs,omitempty"`

	// Deprecated: use Remotes
	Remote  *Remote   `mapstructure:"remote,omitempty"`
	Remotes []*Remote `mapstructure:"remotes,omitempty"`

	Hooks map[string]*Hook `mapstructure:"-"`
}

func (c *Config) Validate() error {
	return version.CheckCovered(c.MinVersion)
}

func (c *Config) Md5() (checksum string, err error) {
	configBytes := new(bytes.Buffer)

	err = c.Dump(JSONCompactFormat, configBytes)
	if err != nil {
		return
	}

	hash := md5.New()
	_, err = io.Copy(hash, configBytes)
	if err != nil {
		return
	}

	checksum = hex.EncodeToString(hash.Sum(nil)[:16])
	return
}

func (c *Config) Dump(format DumpFormat, out io.Writer) error {
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
		dumper = jsonDumper{pretty: true}
	case JSONCompactFormat:
		dumper = jsonDumper{pretty: false}
	default:
		dumper = yamlDumper{}
	}

	return dumper.Dump(res, out)
}

type dumper interface {
	Dump(map[string]interface{}, io.Writer) error
}

type yamlDumper struct{}

func (yamlDumper) Dump(input map[string]interface{}, out io.Writer) error {
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(yamlIndent)
	defer encoder.Close()

	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}

type tomlDumper struct{}

func (tomlDumper) Dump(input map[string]interface{}, out io.Writer) error {
	encoder := toml.NewEncoder(out)
	err := encoder.Encode(input)
	if err != nil {
		return err
	}

	return nil
}

type jsonDumper struct {
	pretty bool
}

func (j jsonDumper) Dump(input map[string]interface{}, out io.Writer) error {
	var res []byte
	var err error
	if j.pretty {
		res, err = json.MarshalIndent(input, "", "  ")
	} else {
		res, err = json.Marshal(input)
	}
	if err != nil {
		return err
	}

	n, err := out.Write(res)
	if n != len(res) {
		return fmt.Errorf("file not written fully: %d/%d", n, len(res))
	}
	if err != nil {
		return err
	}

	if j.pretty {
		_, _ = out.Write([]byte("\n"))
	}

	return nil
}
