package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"path/filepath"
	"strings"
)

// Loads configs from the given directory with extensions
func Load(fs afero.Fs, path string) (*Config, error) {
	global, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	extends, err := readExtends(fs, path, global)
	if err != nil {
		return nil, err
	}

	var config Config

	config.Colors = true // by default colors are enabled

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

func readExtends(fs afero.Fs, path string, global *viper.Viper) (*viper.Viper, error) {
	local, err := read(fs, path, "lefthook-local")
	if err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	// Merge and extend configs
	var extends *viper.Viper

	if local != nil {
		extends = local

		err := extend(fs, extends, local)
		if err != nil {
			return nil, err
		}
	}

	err = extend(fs, extends, global)
	if err != nil {
		return nil, err
	}

	return extends, nil
}

func extend(fs afero.Fs, dest, src *viper.Viper) error {
	for _, path := range src.GetStringSlice("extends") {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		another, err := read(fs, filepath.Dir(path), name)
		if err != nil {
			return err
		}
		dest.MergeConfigMap(another.AllSettings())
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
	base.MergeConfigMap(extra.AllSettings())
	if err := base.Unmarshal(c); err != nil {
		return err
	}

	return nil
}
