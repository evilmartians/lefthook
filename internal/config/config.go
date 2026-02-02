package config

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
	"go.yaml.in/yaml/v3"
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
	MinVersion string `json:"min_version,omitempty" jsonschema:"description=Specify a minimum version for the lefthook binary" koanf:"min_version" mapstructure:"min_version,omitempty"`

	Lefthook string `json:"lefthook,omitempty" jsonschema:"description=Lefthook executable path or command" mapstructure:"lefthook,omitempty"`

	SourceDir string `json:"source_dir,omitempty" jsonschema:"default=.lefthook/,description=Change a directory for script files. Directory for script files contains folders with git hook names which contain script files." koanf:"source_dir" mapstructure:"source_dir,omitempty"`

	SourceDirLocal string `json:"source_dir_local,omitempty" jsonschema:"default=.lefthook-local/,description=Change a directory for local script files (not stored in VCS)" koanf:"source_dir_local" mapstructure:"source_dir_local,omitempty"`

	Rc string `json:"rc,omitempty" jsonschema:"description=Provide an rc file - a simple sh script" mapstructure:"rc,omitempty"`

	Output any `json:"output,omitempty" jsonschema:"oneof_type=boolean;array,description=Manage verbosity by skipping the printing of output of some steps" mapstructure:"output,omitempty"`

	Colors any `json:"colors,omitempty" jsonschema:"description=Enable disable or set your own colors for lefthook output,default=true,oneof_type=boolean;object" mapstructure:"colors,omitempty"`

	Extends []string `json:"extends,omitempty" jsonschema:"description=Specify files to extend config with" mapstructure:"extends,omitempty"`

	NoTTY bool `json:"no_tty,omitempty" jsonschema:"description=Whether hide spinner and other interactive things" koanf:"no_tty" mapstructure:"no_tty,omitempty"`

	AssertLefthookInstalled bool `json:"assert_lefthook_installed,omitempty" koanf:"assert_lefthook_installed" mapstructure:"assert_lefthook_installed,omitempty"`

	SkipLFS bool `json:"skip_lfs,omitempty" jsonschema:"description=Skip running Git LFS hooks (enabled by default)" koanf:"skip_lfs" mapstructure:"skip_lfs,omitempty"`

	NoAutoInstall bool `json:"no_auto_install,omitempty" jsonschema:"description=Do not automatically install hooks when running lefthook" koanf:"no_auto_install" mapstructure:"no_auto_install,omitempty"`

	InstallNonGitHooks bool `json:"install_non_git_hooks,omitempty" jsonschema:"description=Install non-Git hooks to .git/hooks" koanf:"install_non_git_hooks" mapstructure:"install_non_git_hooks,omitempty"`

	GlobMatcher string `json:"glob_matcher,omitempty" jsonschema:"description=Choose the glob matching engine: 'gobwas' (default) or 'doublestar' (standard ** behavior),enum=gobwas,enum=doublestar,default=gobwas" koanf:"glob_matcher" mapstructure:"glob_matcher,omitempty"`

	Remotes []*Remote `json:"remotes,omitempty" jsonschema:"description=Provide multiple remote configs to use lefthook configurations shared across projects. Lefthook will automatically download and merge configurations into main config." mapstructure:"remotes,omitempty"`

	Templates map[string]string `json:"templates,omitempty" jsonschema:"description=Custom templates for replacements in run commands." mapstructure:"templates,omitempty"`

	Hooks map[string]*Hook `jsonschema:"-" mapstructure:"-"`
}

func (c *Config) Md5() (checksum string, err error) {
	configBytes := new(bytes.Buffer)

	err = c.Dump(JSONCompactFormat, configBytes)
	if err != nil {
		return checksum, err
	}

	hash := md5.New()
	_, err = io.Copy(hash, configBytes)
	if err != nil {
		return checksum, err
	}

	checksum = hex.EncodeToString(hash.Sum(nil)[:16])
	return checksum, err
}

func (c *Config) Dump(format DumpFormat, out io.Writer) error {
	res := make(map[string]any)
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
	Dump(map[string]any, io.Writer) error
}

type yamlDumper struct{}

func (yamlDumper) Dump(input map[string]any, out io.Writer) error {
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(yamlIndent)

	err := errors.Join(encoder.Encode(input), encoder.Close())
	return err
}

type tomlDumper struct{}

func (tomlDumper) Dump(input map[string]any, out io.Writer) error {
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

func (j jsonDumper) Dump(input map[string]any, out io.Writer) error {
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
