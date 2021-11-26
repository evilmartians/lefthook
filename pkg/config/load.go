package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"path/filepath"
	"strings"

	"encoding/json"
	"fmt"
)

// Loads configs from the given directory
func Load(fs afero.Fs, path string) (*Config, error) {
	// Read raw config data
	globalViper, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	localViper, err := read(fs, path, "lefthook-local")
	if err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	// Merge and extend configs
	extraViper, err := extendVipers(fs, globalViper, localViper)

	var config Config

	err = unmarshalConfigs(globalViper, extraViper, &config)
	if err != nil {
		return nil, err
	}
	j, _ := json.Marshal(config)
	fmt.Printf(string(j))
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

func extendVipers(fs afero.Fs, dest, src *viper.Viper) (*viper.Viper, error) {
	if src == nil {
		src = viper.New()
	}

	// TODO: Refactor
	for _, path := range dest.GetStringSlice("extends") {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		another, err := read(fs, filepath.Dir(path), name)
		if err != nil {
			return nil, err
		}
		src.MergeConfigMap(another.AllSettings())
	}

	for _, path := range src.GetStringSlice("extends") {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		another, err := read(fs, filepath.Dir(path), name)
		if err != nil {
			return nil, err
		}
		src.MergeConfigMap(another.AllSettings())
	}

	return src, nil
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
