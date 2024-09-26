package config

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	koanffs "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/afero"

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
	names            = []string{"lefthook", ".lefthook"}
	extensions       = []string{".yml", ".yaml", ".toml", ".json"}
	parsers          = map[string]func() koanf.Parser{
		".json": jsonParser,
		".yaml": yamlParser,
		".yml":  yamlParser,
		".toml": tomlParser,
	}
)

type ErrNotFound struct {
	msg string
}

func (e ErrNotFound) Error() string {
	return e.msg
}

func yamlParser() koanf.Parser {
	return yaml.Parser()
}

func jsonParser() koanf.Parser {
	return json.Parser()
}

func tomlParser() koanf.Parser {
	return toml.Parser()
}

type fs struct {
	fs afero.Fs
}

func (f fs) Open(name string) (iofs.File, error) {
	return f.fs.Open(name)
}

// Loads configs from the given directory with extensions.
func Load(f afero.Fs, repo *git.Repository) (*Config, error) {
	fs := fs{fs: f}
	global, err := readOne(fs, repo.RootPath)
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

func read(fs fs, path string, name string, parser func() koanf.Parser) (*koanf.Koanf, error) {
	k := koanf.New(".")
	if err := k.Load(koanffs.Provider(fs, filepath.Join(path, name)), parser()); err != nil {
		return nil, err
	}

	return k, nil
}

func readOne(fs fs, path string) (*koanf.Koanf, error) {
	for _, name := range names {
		for _, ext := range extensions {
			k, err := read(fs, path, name+ext, parsers[ext])
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}

				return nil, err
			}

			return k, nil
		}
	}

	return nil, ErrNotFound{fmt.Sprintf("No config files were found in \"%s\"", path)}
}

// mergeAll merges configs using the following order.
// - lefthook/.lefthook
// - files from `extends`
// - files from `remotes`
// - lefthook-local/.lefthook-local.
func mergeAll(fs fs, repo *git.Repository) (*koanf.Koanf, error) {
	extends, err := readOne(fs, repo.RootPath)
	if err != nil {
		return nil, err
	}

	if err := extend(fs, extends, repo.RootPath); err != nil {
		return nil, err
	}

	// Save global extends to compare them after merging local config
	globalExtends := extends.Strings("extends")

	if err := mergeRemotes(fs, repo, extends); err != nil {
		return nil, err
	}

	//nolint:nestif
	if err := mergeLocal(fs, extends, repo.RootPath); err == nil {
		// Local extends need to be re-applied only if they have different settings
		localExtends := extends.Strings("extends")
		if !slices.Equal(globalExtends, localExtends) {
			if err = extend(fs, extends, repo.RootPath); err != nil {
				return nil, err
			}
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return extends, nil
}

// mergeRemotes merges remote configs to the current one.
func mergeRemotes(fs fs, repo *git.Repository, k *koanf.Koanf) error {
	var remote *Remote // Deprecated
	var remotes []*Remote

	err := k.Unmarshal("remotes", &remotes)
	if err != nil {
		return err
	}

	// Deprecated
	err = k.Unmarshal("remote", &remote)
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

			// _, err = fs.Stat(configPath)
			// if err != nil {
			// 	continue
			// }

			if err = merge(fs, k, configPath); err != nil {
				return err
			}

			if err = extend(fs, k, filepath.Dir(configPath)); err != nil {
				return err
			}
		}

		// Reset extends to omit issues when extending with remote extends.
		err = k.Set("extends", nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// extend merges all files listed in 'extends' option into the config.
func extend(fs fs, k *koanf.Koanf, root string) error {
	return extendRecursive(fs, k, root, make(map[string]struct{}))
}

// extendRecursive merges extends.
// If extends contain other extends they get merged too.
func extendRecursive(fs fs, k *koanf.Koanf, root string, extends map[string]struct{}) error {
	for _, path := range k.Strings("extends") {
		if _, contains := extends[path]; contains {
			return fmt.Errorf("possible recursion in extends: path %s is specified multiple times", path)
		}
		extends[path] = struct{}{}

		ext := filepath.Ext(path)
		if len(ext) == 0 || parsers[ext] == nil {
			return fmt.Errorf("unable to parse an extension: %s", path)
		}

		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}

		ko := koanf.New(".")
		if err := ko.Load(koanffs.Provider(fs, path), parsers[ext]()); err != nil {
			return fmt.Errorf("failed to load file %s: %w", path, err)
		}

		if err := extendRecursive(fs, ko, root, extends); err != nil {
			return err
		}

		if err := k.Merge(ko); err != nil {
			return err
		}
	}

	return nil
}

// merge merges the configuration with a new one parsed from `path`.
func merge(fs fs, k *koanf.Koanf, path string) error {
	ext := filepath.Ext(path)
	if len(ext) == 0 || parsers[ext] == nil {
		return fmt.Errorf("unable to parse an extension: %s", path)
	}

	ko := koanf.New(".")
	if err := ko.Load(koanffs.Provider(fs, path), parsers[ext]()); err != nil {
		return fmt.Errorf("failed to load file %s: %w", path, err)
	}

	if err := k.Merge(ko); err != nil {
		return err
	}

	return nil
}

// mergeLocal merges local configurations if they exist.
func mergeLocal(fs fs, k *koanf.Koanf, root string) error {
	for _, name := range localConfigNames {
		for _, ext := range extensions {
			err := merge(fs, k, filepath.Join(root, name+ext))
			if err == nil {
				break
			}

			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}

	return nil
}

func unmarshalConfigs(base, extra *koanf.Koanf, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for _, hookName := range AvailableHooks {
		if err := addHook(hookName, base, extra, c); err != nil {
			return err
		}
	}

	// For extra non-git hooks.
	// This behavior may be deprecated in next versions.
	// Notice that with append we're allowing extra hooks to be added in local config
	for _, maybeHook := range append(base.Keys(), extra.Keys()...) {
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
	if err := base.Merge(extra); err != nil {
		return err
	}

	if err := base.Unmarshal(".", c); err != nil {
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

func addHook(hookName string, base, extra *koanf.Koanf, c *Config) error {
	if !extra.Exists(hookName) {
		return nil
	}

	if !base.Exists(hookName) {
		base.Set(hookName, extra.Cut(hookName).Raw())
		return nil
	}

	baseHook := base.Cut(hookName)
	extraHook := extra.Cut(hookName)

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
