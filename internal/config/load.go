package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
)

const (
	DefaultConfigName     = "lefthook.yml"
	DefaultSourceDir      = ".lefthook"
	DefaultSourceDirLocal = ".lefthook-local"
	DefaultColorsEnabled  = true
)

var hookKeyRegexp = regexp.MustCompile(`^(?P<hookName>[^.]+)\.(scripts|commands)`)

// NotFoundError wraps viper.ConfigFileNotFoundError for lefthook.
type NotFoundError struct {
	message string
}

// Error returns message of viper.ConfigFileNotFoundError.
func (err NotFoundError) Error() string {
	return err.message
}

// Loads configs from the given directory with extensions.
func Load(fs afero.Fs, repo *git.Repository) (*Config, error) {
	global, err := read(fs, repo.RootPath, "lefthook")
	if err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if ok := errors.As(err, &notFoundErr); ok {
			return nil, NotFoundError{err.Error()}
		}

		return nil, err
	}

	extends, err := mergeAll(fs, repo)
	if err != nil {
		return nil, err
	}

	var config Config

	config.Colors = DefaultColorsEnabled
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

// mergeAll merges remotes and extends from .lefthook and .lefthook-local.
func mergeAll(fs afero.Fs, repo *git.Repository) (*viper.Viper, error) {
	extends, err := read(fs, repo.RootPath, "lefthook")
	if err != nil {
		return nil, err
	}

	if err := extend(extends, repo.RootPath); err != nil {
		return nil, err
	}

	if err := mergeRemote(fs, repo, extends); err != nil {
		return nil, err
	}

	if err := merge("lefthook-local", "", extends); err == nil {
		if err = extend(extends, repo.RootPath); err != nil {
			return nil, err
		}
	} else {
		var notFoundErr viper.ConfigFileNotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			return nil, err
		}
	}

	return extends, nil
}

// mergeRemote merges remote config to the current one.
func mergeRemote(fs afero.Fs, repo *git.Repository, v *viper.Viper) error {
	var remote Remote
	err := v.UnmarshalKey("remote", &remote)
	if err != nil {
		return err
	}

	if !remote.Configured() {
		return nil
	}

	remotePath := repo.RemoteFolder(remote.GitURL)
	configFile := DefaultConfigName
	if len(remote.Config) > 0 {
		configFile = remote.Config
	}
	configPath := filepath.Join(remotePath, configFile)

	log.Debugf("Merging remote config: %s", configPath)

	_, err = fs.Stat(configPath)
	if err != nil {
		return nil
	}

	if err := merge("remote", configPath, v); err != nil {
		return err
	}

	if err := extend(v, filepath.Dir(configPath)); err != nil {
		return err
	}

	return nil
}

// extend merges all files listed in 'extends' option into the config.
func extend(v *viper.Viper, root string) error {
	log.Debugf("extends %v\n", v.GetStringSlice("extends"))
	for i, path := range v.GetStringSlice("extends") {
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if err := merge(fmt.Sprintf("extend_%d", i), path, v); err != nil {
			return err
		}
	}
	return nil
}

// merge merges the configuration using viper builtin MergeInConfig.
func merge(name, path string, v *viper.Viper) error {
	v.SetConfigName(name)
	if len(path) > 0 {
		v.SetConfigFile(path)
	}
	if err := v.MergeInConfig(); err != nil {
		return err
	}

	return nil
}

func unmarshalConfigs(base, extra *viper.Viper, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for _, hookName := range AvailableHooks {
		if err := addHook(hookName, base, extra, c); err != nil {
			return err
		}
	}

	// For extra non-git hooks.
	// This behavior may be deprecated in next versions.
	// Notice that with append we're allowing extra hooks to be added in local config
	for _, maybeHook := range append(base.AllKeys(), extra.AllKeys()...) {
		if !hookKeyRegexp.MatchString(maybeHook) {
			continue
		}

		matches := hookKeyRegexp.FindStringSubmatch(maybeHook)
		hookName := matches[hookKeyRegexp.SubexpIndex("hookName")]
		if _, ok := c.Hooks[hookName]; ok {
			continue
		}

		if err := addHook(hookName, base, extra, c); err != nil {
			return err
		}
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

func addHook(hookName string, base, extra *viper.Viper, c *Config) error {
	baseHook := base.Sub(hookName)
	extraHook := extra.Sub(hookName)

	resultHook, err := unmarshalHooks(baseHook, extraHook)
	if err != nil {
		return err
	}

	if resultHook == nil {
		return nil
	}

	c.Hooks[hookName] = resultHook

	return nil
}
