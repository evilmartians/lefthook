package config

import (
	"github.com/spf13/viper"
	"path/filepath"
	"strings"

	git "github.com/evilmartians/lefthook/pkg/git"
)

const (
	runnerWrapPattern = "{cmd}"
)

type Config struct {
	Extras struct {
		Colors         bool     `mapstructure:"colors"`
		MinVersion     string   `mapstructure:"min_version"`
		SkipOutput     []string `mapstructure:"skip_output"`
		SourceDir      string   `mapstructure:"source_dir"`
		SourceDirLocal string   `mapstructure:"source_dir_local"`
	}

	Hooks map[string]*Hook
}

// Loads configs from the given directory
func Load(path string) (*Config, error) {
	var config Config

	globalViper := newViper(path, "lefthook")
	localViper := newViper(path, "lefthook-local")

	// Read the global config
	if err := globalViper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := localViper.MergeConfigMap(globalViper.AllSettings()); err != nil {
		return nil, err
	}

	// Read and merge lefthook-local
	if err := localViper.MergeInConfig(); err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	// Merge all extensions if specified
	if err := extendConfig(localViper); err != nil {
		return nil, err
	}

	// Allow overwriting settings with ENV variables
	localViper.SetEnvPrefix("LEFTHOOK")
	localViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	localViper.AutomaticEnv()

	// Unmarshal extra part of the config into struct.
	// Hooks are going to be unmarshalled in a lazy way just not
	// to waste time on parsing possibly unneded data.
	if err := localViper.Unmarshal(&config.Extras); err != nil {
		return nil, err
	}

	// Note could be done in a lazy way but makes more sense when explicit
	if err := unmarshalHooks(localViper, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func newViper(path, name string) *viper.Viper {
	c := viper.New()
	c.AddConfigPath(path)
	c.SetConfigName(name)
	return c
}

func unmarshalHooks(v *viper.Viper, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for _, hookName := range git.AvailableHooks {
		hookConfig := v.Sub(hookName)
		if hookConfig == nil {
			continue
		}
		var hook Hook
		if err := hookConfig.Unmarshal(&hook); err != nil {
			return err
		}

		c.Hooks[hookName] = &hook
	}

	return nil
}

// Handle `extends` setting that merges many lefthook files into one
func extendConfig(local *viper.Viper) error {
	extends := local.GetStringSlice("extends")
	if len(extends) == 0 {
		return nil
	}

	for _, path := range extends {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		local.SetConfigName(name)
		local.AddConfigPath(filepath.Dir(path))

		if err := local.MergeInConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) Hook(name string, v *viper.Viper) (*Hook, error) {
	hook, err := extractHook(name, v)
	if err != nil {
		return hook, err
	}

	return hook, nil
}

func extractHook(name string, v *viper.Viper) (*Hook, error) {
	config := v.Sub(name)
	if config == nil {
		return nil, nil
	}

	hook := &Hook{}
	if err := config.Unmarshal(hook); err != nil {
		return nil, err
	}
	return hook, nil
}
