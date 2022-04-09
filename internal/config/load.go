package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	DefaultSourceDir      = ".lefthook"
	DefaultSourceDirLocal = ".lefthook-local"
)

// Loads configs from the given directory with extensions.
func Load(fs afero.Fs, path string) (*Config, error) {
	global, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	extends, err := mergeAllExtends(fs, path)
	if err != nil {
		return nil, err
	}

	var config Config

	config.Colors = true // by default colors are enabled
	config.SourceDir = DefaultSourceDir
	config.SourceDirLocal = DefaultSourceDirLocal

	err = unmarshalConfigs(global, extends, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func read(fs afero.Fs, path string, name string) (*viper.Viper, error) {
	v := viper.New()
	v.SetFs(fs)
	v.AddConfigPath(path)
	v.SetConfigName(name)

	// Allow overwriting settings with ENV variables
	v.SetEnvPrefix("LEFTHOOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}

// Merges extends from .lefthook and .lefthook-local.
func mergeAllExtends(fs afero.Fs, path string) (*viper.Viper, error) {
	extends, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	if err := extend(fs, extends); err != nil {
		return nil, err
	}

	extends.SetConfigName("lefthook-local")
	if err := extends.MergeInConfig(); err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	if err := extend(fs, extends); err != nil {
		return nil, err
	}

	return extends, nil
}

func extend(fs afero.Fs, v *viper.Viper) error {
	for _, path := range v.GetStringSlice("extends") {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		another, err := read(fs, filepath.Dir(path), name)
		if err != nil {
			return err
		}
		if err = v.MergeConfigMap(another.AllSettings()); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalConfigs(base, extra *viper.Viper, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for _, hookName := range AvailableHooks {
		baseHook := base.Sub(hookName)
		extraHook := extra.Sub(hookName)

		resultHook, err := unmarshalHooks(baseHook, extraHook)
		if err != nil {
			return err
		}

		if resultHook == nil {
			continue
		}

		resultHook.processDeprecations()

		c.Hooks[hookName] = resultHook
	}

	// Merge config and unmarshal it
	if err := base.MergeConfigMap(extra.AllSettings()); err != nil {
		return err
	}
	if err := base.Unmarshal(c); err != nil {
		return err
	}

	return nil
}
