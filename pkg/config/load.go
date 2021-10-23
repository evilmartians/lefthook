package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"path/filepath"
	"strings"
)

// Loads configs from the given directory
func Load(fs afero.Fs, path string) (*Config, error) {
	globalConfig, err := parseConfig(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	localConfig, err := parseConfig(fs, path, "lefthook-local")
	if err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	if localConfig != nil {
		globalConfig.Merge(localConfig)
	}

	err = extendConfig(fs, globalConfig)

	return globalConfig, err
}

func parseConfig(fs afero.Fs, path string, name string) (*Config, error) {
	v := viper.New()
	v.SetFs(fs)
	v.AddConfigPath(path)
	v.SetConfigName(name)

	// Allow overwriting settings with ENV variables
	v.SetEnvPrefix("LEFTHOOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var config Config

	// Set defaults
	config.Colors = true

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal extra part of the config into struct.
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Note could be done in a lazy way but makes more sense when explicit
	if err := unmarshalHooks(v, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func unmarshalHooks(v *viper.Viper, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for _, hookName := range AvailableHooks {
		hookConfig := v.Sub(hookName)
		if hookConfig == nil {
			continue
		}

		var hook Hook
		if err := unmarshalHook(hookConfig, &hook); err != nil {
			return err
		}

		hook.processDeprecations()

		c.Hooks[hookName] = &hook
	}

	return nil
}

func extendConfig(fs afero.Fs, c *Config) error {
	if len(c.Extends) == 0 {
		return nil
	}

	for _, path := range c.Extends {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		another, err := parseConfig(fs, filepath.Dir(path), name)
		if err != nil {
			return err
		}

		c.Merge(another)
	}

	return nil
}
