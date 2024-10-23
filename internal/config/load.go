package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
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
)

var (
	hookKeyRegexp    = regexp.MustCompile(`^(?P<hookName>[^.]+)\.(scripts|commands)`)
	localConfigNames = []string{"lefthook-local", ".lefthook-local"}
)

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
	global, err := readOne(fs, repo.RootPath, []string{"lefthook", ".lefthook"})
	if err != nil {
		return nil, err
	}

	extends, err := mergeAll(fs, repo)
	if err != nil {
		return nil, err
	}

	var config Config

	config.SourceDir = DefaultSourceDir
	config.SourceDirLocal = DefaultSourceDirLocal

	err = unmarshalConfigs(global, extends, &config)
	if err != nil {
		return nil, err
	}

	log.SetColors(config.Colors)
	return &config, nil
}

func read(fs afero.Fs, path string, name string) (*viper.Viper, error) {
	v := newViper(fs, path)
	v.SetConfigName(name)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}

func newViper(fs afero.Fs, path string) *viper.Viper {
	v := viper.New()
	v.SetFs(fs)
	v.AddConfigPath(path)

	// Allow overwriting settings with ENV variables
	v.SetEnvPrefix("LEFTHOOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v
}

func readOne(fs afero.Fs, path string, names []string) (*viper.Viper, error) {
	for _, name := range names {
		v, err := read(fs, path, name)
		if err != nil {
			var notFoundErr viper.ConfigFileNotFoundError
			if ok := errors.As(err, &notFoundErr); ok {
				continue
			}

			return nil, err
		}

		return v, nil
	}

	return nil, NotFoundError{fmt.Sprintf("No config files with names %q have been found in \"%s\"", names, path)}
}

// mergeAll merges configs using the following order.
// - lefthook/.lefthook
// - files from `extends`
// - files from `remotes`
// - lefthook-local/.lefthook-local.
func mergeAll(fs afero.Fs, repo *git.Repository) (*viper.Viper, error) {
	extends, err := readOne(fs, repo.RootPath, []string{"lefthook", ".lefthook"})
	if err != nil {
		return nil, err
	}

	if err := extend(fs, extends, repo.RootPath); err != nil {
		return nil, err
	}

	// Save global extends to compare them after merging local config
	globalExtends := extends.GetStringSlice("extends")

	if err := mergeRemotes(fs, repo, extends); err != nil {
		return nil, err
	}

	//nolint:nestif
	if err := mergeLocal(extends); err == nil {
		// Local extends need to be re-applied only if they have different settings
		localExtends := extends.GetStringSlice("extends")
		if !slices.Equal(globalExtends, localExtends) {
			if err = extend(fs, extends, repo.RootPath); err != nil {
				return nil, err
			}
		}
	} else {
		var notFoundErr viper.ConfigFileNotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			return nil, err
		}
	}

	return extends, nil
}

// mergeRemotes merges remote configs to the current one.
func mergeRemotes(fs afero.Fs, repo *git.Repository, v *viper.Viper) error {
	var remote *Remote // Deprecated
	var remotes []*Remote

	err := v.UnmarshalKey("remotes", &remotes)
	if err != nil {
		return err
	}

	// Deprecated
	err = v.UnmarshalKey("remote", &remote)
	if err != nil {
		return err
	}

	// Backward compatibility
	if remote != nil {
		remotes = append(remotes, remote)
	}

	for _, remote := range remotes {
		if !remote.Configured() {
			continue
		}

		// Use for backward compatibility with "remote(s).config"
		if remote.Config != "" {
			remote.Configs = append(remote.Configs, remote.Config)
		}

		if len(remote.Configs) == 0 {
			remote.Configs = append(remote.Configs, DefaultConfigName)
		}

		for _, config := range remote.Configs {
			remotePath := repo.RemoteFolder(remote.GitURL, remote.Ref)
			configFile := config
			configPath := filepath.Join(remotePath, configFile)

			log.Debugf("Merging remote config: %s: %s", remote.GitURL, configPath)

			_, err = fs.Stat(configPath)
			if err != nil {
				continue
			}

			if err = merge("remotes", configPath, v); err != nil {
				return err
			}

			if err = extend(fs, v, filepath.Dir(configPath)); err != nil {
				return err
			}
		}

		// Reset extends to omit issues when extending with remote extends.
		err = v.MergeConfigMap(map[string]interface{}{"extends": nil})
		if err != nil {
			return err
		}
	}

	return nil
}

// extend merges all files listed in 'extends' option into the config.
func extend(fs afero.Fs, v *viper.Viper, root string) error {
	return extendRecursive(fs, v, root, make(map[string]struct{}))
}

// extendRecursive merges extends.
// If extends contain other extends they get merged too.
func extendRecursive(fs afero.Fs, v *viper.Viper, root string, extends map[string]struct{}) error {
	for _, pathOrGlob := range v.GetStringSlice("extends") {
		if !filepath.IsAbs(pathOrGlob) {
			pathOrGlob = filepath.Join(root, pathOrGlob)
		}

		paths, err := afero.Glob(fs, pathOrGlob)
		if err != nil {
			return fmt.Errorf("bad glob syntax for '%s': %w", pathOrGlob, err)
		}

		for _, path := range paths {
			if _, contains := extends[path]; contains {
				return fmt.Errorf("possible recursion in extends: path %s is specified multiple times", path)
			}
			extends[path] = struct{}{}

			extendV := newViper(fs, root)
			extendV.SetConfigFile(path)
			if err := extendV.ReadInConfig(); err != nil {
				return err
			}

			if err := extendRecursive(fs, extendV, root, extends); err != nil {
				return err
			}

			if err := v.MergeConfigMap(extendV.AllSettings()); err != nil {
				return err
			}
		}
	}

	return nil
}

// merge merges the configuration using viper builtin MergeInConfig.
func merge(name, path string, v *viper.Viper) error {
	v.SetConfigName(name)
	v.SetConfigFile(path)
	return v.MergeInConfig()
}

func mergeLocal(v *viper.Viper) error {
	for _, name := range localConfigNames {
		err := merge(name, "", v)
		if err == nil {
			break
		}

		var notFoundErr viper.ConfigFileNotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			return err
		}
	}

	return nil
}

func unmarshalConfigs(base, extra *viper.Viper, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for hookName := range AvailableHooks {
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

	// Deprecation handling

	if c.Remote != nil {
		log.Warn("DEPRECATED: \"remote\" option is deprecated and will be omitted in the next major release, use \"remotes\" option instead")
		c.Remotes = append(c.Remotes, c.Remote)
	}
	c.Remote = nil

	for _, remote := range c.Remotes {
		if remote.Config != "" {
			log.Warn("DEPRECATED: \"remotes\".\"config\" option is deprecated and will be omitted in the next major release, use \"configs\" option instead")
			remote.Configs = append(remote.Configs, remote.Config)
		}

		remote.Config = ""
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
