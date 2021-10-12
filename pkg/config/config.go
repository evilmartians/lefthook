package config

import (
	"github.com/spf13/viper"
	"path/filepath"
)

type Config struct {
	base *viper.Viper
	full *viper.Viper
}

const (
	runnerWrapPattern = "{cmd}"
)

func Load(path string) (*Config, error) {
	config := &Config{
		base: newViper(path, "lefthook"),
		full: newViper(path, "lefthook-local"),
	}
	if err := config.base.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := config.full.MergeConfigMap(config.base.AllSettings()); err != nil {
		return nil, err
	}
	if err := config.full.MergeInConfig(); err != nil {
		return nil, err
	}
	if err := extendConfig(config); err != nil {
		return nil, err
	}
	// config.full.SetEnvPrefix("LEFTHOOK") // TODO: uncomment?
	config.full.AutomaticEnv()

	return config, nil
}

func extendConfig(config *Config) error {
	extends := config.full.GetStringSlice("extends")
	if len(extends) > 0 {
		for _, path := range extends {
			filename := filepath.Base(path)
			extension := filepath.Ext(path)
			name := filename[0 : len(filename)-len(extension)]
			config.full.SetConfigName(name)
			config.full.AddConfigPath(filepath.Dir(path))
			if err := config.full.MergeInConfig(); err != nil {
				return err
			}
		}
	}
	return nil
}

func newViper(path, name string) *viper.Viper {
	c := viper.New()
	c.AddConfigPath(path)
	c.SetConfigName(name)
	return c
}

func (c *Config) Main() (*MainConfig, error) {
	mainConfig := NewMainConfig()
	if c.full != nil {
		if err := c.full.Unmarshal(mainConfig); err != nil {
			return nil, err
		}
	}
	return mainConfig, nil
}

func (c *Config) Hook(name string) (*Hook, error) {
	hook, err := extractHook(name, c.full)
	if hook == nil {
		return hook, err
	}

	baseHook, err := extractHook(name, c.base)
	if baseHook != nil {
		hook.expandWith(baseHook)
	}

	return hook, nil
}

func extractHook(name string, full *viper.Viper) (*Hook, error) {
	config := full.Sub(name)
	if config == nil {
		return nil, nil
	}

	hook := &Hook{}
	if err := config.Unmarshal(hook); err != nil {
		return nil, err
	}
	return hook, nil
}
