// TODO rewrite using Koanf
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	kfs "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
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
	mainConfigNames  = []string{"lefthook", ".lefthook"}
	extensions       = []string{
		".yml",
		".yaml",
		".json",
		".toml",
	}
	parsers = map[string]koanf.Parser{
		".yml":  yaml.Parser(),
		".yaml": yaml.Parser(),
		".json": json.Parser(),
		".toml": toml.Parser(),
	}
)

// NotFoundError wraps viper.ConfigFileNotFoundError for lefthook.
type NotFoundError struct {
	message string
}

// Error returns message of viper.ConfigFileNotFoundError.
func (err NotFoundError) Error() string {
	return err.message
}

func loadOne(k *koanf.Koanf, filesystem afero.Fs, root string, names []string) error {
	for _, extension := range extensions {
		for _, name := range names {
			config := filepath.Join(root, name+extension)
			if ok, _ := afero.Exists(filesystem, config); !ok {
				continue
			}

			if err := k.Load(kfs.Provider(afero.NewIOFS(filesystem), config), parsers[extension]); err != nil {
				return err
			}

			return nil
		}
	}

	return NotFoundError{fmt.Sprintf("No config files with names %q have been found in \"%s\"", names, root)}
}

// Loads configs from the given directory with extensions.
func Load(filesystem afero.Fs, repo *git.Repository) (*Config, error) {
	main := koanf.New(".")

	// Load main (e.g. lefthook.yml)
	if err := loadOne(main, filesystem, repo.RootPath, mainConfigNames); err != nil {
		return nil, err
	}

	// Save `extends` and `remotes`
	extends := main.Strings("extends")
	var remotes []*Remote
	if err := main.Unmarshal("remotes", &remotes); err != nil {
		return nil, err
	}

	// Deprecated
	var remote *Remote
	if err := main.Unmarshal("remote", &remote); err != nil {
		return nil, err
	}

	// Backward compatibility for `remote`. Will be deleted in future major release
	if remote != nil {
		remotes = append(remotes, remote)
	}

	secondary := koanf.New(".")

	// Load main `extends`
	if err := extend(secondary, filesystem, repo.RootPath, extends); err != nil {
		return nil, err
	}

	// Load main `remotes`
	if err := loadRemotes(secondary, filesystem, repo, remotes); err != nil {
		return nil, err
	}

	// Load optional local config (e.g. lefthook-local.yml)
	if err := loadOne(secondary, filesystem, repo.RootPath, localConfigNames); err != nil {
		var notFoundErr NotFoundError
		if ok := errors.As(err, &notFoundErr); !ok {
			return nil, err
		}
	}

	var config Config

	config.SourceDir = DefaultSourceDir
	config.SourceDirLocal = DefaultSourceDirLocal

	// TODO: continue here, merge rules must be applied
	if err := unmarshalConfigs(main, secondary, &config); err != nil {
		return nil, err
	}

	log.SetColors(config.Colors)
	return &config, nil
}

// func read(fs afero.Fs, name, path string) (*viper.Viper, error) {
// 	v := newViper(fs, path)
// 	v.SetConfigName(name)

// 	if err := v.ReadInConfig(); err != nil {
// 		return nil, err
// 	}

// 	return v, nil
// }

// func newViper(fs afero.Fs, path string) *viper.Viper {
// 	v := viper.New()
// 	v.SetFs(fs)
// 	v.AddConfigPath(path)

// 	// Allow overwriting settings with ENV variables
// 	v.SetEnvPrefix("LEFTHOOK")
// 	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
// 	v.AutomaticEnv()

// 	return v
// }

// loadSecondary reads extends, remotes and local config.
// - files from `extends`
// - files from `remotes`
// - lefthook-local/.lefthook-local.
//func loadSecondary(k *koanf.Koanf, fs afero.Fs, repo *git.Repository, extends []string, remotes []*Remote) (*viper.Viper, error) {
//	//nolint:nestif
//	if err := mergeLocal(secondary); err == nil {
//		// Local extends need to be re-applied only if they have different settings
//		localExtends := secondary.GetStringSlice("extends")
//		if !slices.Equal(extends, localExtends) {
//			if err = extend(fs, repo.RootPath, secondary, localExtends); err != nil {
//				return nil, err
//			}
//		}
//	} else {
//		var notFoundErr viper.ConfigFileNotFoundError
//		if ok := errors.As(err, &notFoundErr); !ok {
//			return nil, err
//		}
//	}

//	return secondary, nil
//}

// loadRemotes merges remote configs to the current one.
func loadRemotes(k *koanf.Koanf, filesystem afero.Fs, repo *git.Repository, remotes []*Remote) error {
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

			if ok, err := afero.Exists(filesystem, configPath); !ok || err != nil {
				continue
			}

			parser, ok := parsers[filepath.Ext(configPath)]
			if !ok {
				panic("TODO: unknown extension to parse")
			}

			if err := k.Load(kfs.Provider(afero.NewIOFS(filesystem), configPath), parser); err != nil {
				return err
			}

			extends := k.Slices("extends")
			if err := extend(k, filesystem, filepath.Dir(configPath), extends); err != nil {
				return err
			}
		}

		// Reset extends to omit issues when extending with remote extends.
		if err := k.Set("extends", nil); err != nil {
			return err
		}
	}

	return nil
}

// extend merges all files listed in 'extends' option into the config.
func extend(k *koanf.Koanf, filesystem afero.Fs, root string, extends []string) error {
	return extendRecursive(k, filesystem, root, extends, make(map[string]struct{}))
}

// extendRecursive merges extends.
// If extends contain other extends they get merged too.
func extendRecursive(k *koanf.Koanf, filesystem afero.Fs, root string, extends []string, visited map[string]struct{}) error {
	for _, pathOrGlob := range extends {
		if !filepath.IsAbs(pathOrGlob) {
			pathOrGlob = filepath.Join(root, pathOrGlob)
		}

		paths, err := afero.Glob(filesystem, pathOrGlob)
		if err != nil {
			return fmt.Errorf("bad glob syntax for '%s': %w", pathOrGlob, err)
		}

		for _, path := range paths {
			if _, contains := visited[path]; contains {
				return fmt.Errorf("possible recursion in extends: path %s is specified multiple times", path)
			}
			visited[path] = struct{}{}

			extent := koanf.New(".")
			parser, ok := parsers[filepath.Ext(path)]
			if !ok {
				panic("TODO: unknown extension for extent " + path)
			}
			if err := extent.Load(kfs.Provider(afero.NewIOFS(filesystem), path), parser); err != nil {
				return err
			}

			if err := extendRecursive(extent, filesystem, root, extent.Strings("extends"), visited); err != nil {
				return err
			}

			if err := k.Merge(extent); err != nil {
				return err
			}
		}
	}

	return nil
}

// // merge merges the configuration using viper builtin MergeInConfig.
// func merge(v *viper.Viper, name, path string) error {
// 	v.SetConfigName(name)
// 	v.SetConfigFile(path)
// 	return v.MergeInConfig()
// }

func loadLocal(k *koanf.Koanf) error {
	for _, name := range localConfigNames {
		if err := merge(v, name, ""); err != nil {
			var notFoundErr viper.ConfigFileNotFoundError
			if ok := errors.As(err, &notFoundErr); ok {
				continue
			}

			return err
		}

		break
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
