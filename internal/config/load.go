package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	kfs "github.com/knadh/koanf/providers/fs"
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
	hookKeyRegexp    = regexp.MustCompile(`^(?P<hookName>[^.]+)\.(?:scripts|commands|jobs)`)
	LocalConfigNames = []string{"lefthook-local", ".lefthook-local", filepath.Join(".config", "lefthook-local")}
	MainConfigNames  = []string{"lefthook", ".lefthook", filepath.Join(".config", "lefthook")}
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

	mergeJobsOption = koanf.WithMergeFunc(mergeJobs)
)

// ConfigNotFoundError.
type ConfigNotFoundError struct {
	message string
}

func (err ConfigNotFoundError) Error() string {
	return err.message
}

func loadOne(k *koanf.Koanf, filesystem afero.Fs, root string, names []string) error {
	for _, extension := range extensions {
		for _, name := range names {
			config := filepath.Join(root, name+extension)
			if ok, _ := afero.Exists(filesystem, config); !ok {
				continue
			}

			log.Debug("loading config: ", config)
			if err := k.Load(kfs.Provider(newIOFS(filesystem), config), parsers[extension], mergeJobsOption); err != nil {
				return err
			}

			return nil
		}
	}

	return ConfigNotFoundError{fmt.Sprintf("No config files with names %q have been found in \"%s\"", names, root)}
}

func LoadKoanf(filesystem afero.Fs, repo *git.Repository) (*koanf.Koanf, *koanf.Koanf, error) {
	main := koanf.New(".")

	// Load main (e.g. lefthook.yml)
	if err := loadOne(main, filesystem, repo.RootPath, MainConfigNames); err != nil {
		return nil, nil, err
	}

	// Save `extends` and `remotes`
	extends := main.Strings("extends")
	var remotes []*Remote
	if err := main.Unmarshal("remotes", &remotes); err != nil {
		return nil, nil, err
	}

	// Deprecated
	var remote *Remote
	if err := main.Unmarshal("remote", &remote); err != nil {
		return nil, nil, err
	}

	// Backward compatibility for `remote`. Will be deleted in future major release
	if remote != nil {
		remotes = append(remotes, remote)
	}

	secondary := koanf.New(".")

	// Load main `extends`
	if err := extend(secondary, filesystem, repo.RootPath, extends); err != nil {
		return nil, nil, err
	}

	// Load main `remotes`
	if err := loadRemotes(secondary, filesystem, repo, remotes); err != nil {
		return nil, nil, err
	}

	// Don't allow to set `lefthook` field from a remote config
	secondary.Delete("lefthook")

	// Load optional local config (e.g. lefthook-local.yml)
	var noLocal bool
	if err := loadOne(secondary, filesystem, repo.RootPath, LocalConfigNames); err != nil {
		var configNotFoundErr ConfigNotFoundError
		if ok := errors.As(err, &configNotFoundErr); !ok {
			return nil, nil, err
		}
		noLocal = true
	}

	// Load local `extends`
	localExtends := secondary.Strings("extends")
	if !noLocal && !slices.Equal(extends, localExtends) {
		if err := extend(secondary, filesystem, repo.RootPath, localExtends); err != nil {
			return nil, nil, err
		}
	}

	return main, secondary, nil
}

// Loads configs from the given directory with extensions.
func Load(filesystem afero.Fs, repo *git.Repository) (*Config, error) {
	main, secondary, err := LoadKoanf(filesystem, repo)
	if err != nil {
		return nil, err
	}

	var config Config

	config.SourceDir = DefaultSourceDir
	config.SourceDirLocal = DefaultSourceDirLocal

	if err := unmarshalConfigs(main, secondary, &config); err != nil {
		return nil, err
	}

	log.SetColors(config.Colors)

	return &config, nil
}

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
				return fmt.Errorf("can't parse config '%[1]s', file has unsupported or no extension\nhint: rename %[1]s to %[1]s.yml", configPath)
			}

			if err := k.Load(kfs.Provider(newIOFS(filesystem), configPath), parser, mergeJobsOption); err != nil {
				return err
			}

			extends := k.Strings("extends")
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
				return fmt.Errorf("can't parse config '%[1]s', file has unsupported or no extension\nhint: rename %[1]s to %[1]s.yml", path)
			}
			if err := extent.Load(kfs.Provider(newIOFS(filesystem), path), parser, mergeJobsOption); err != nil {
				return err
			}

			if err := extendRecursive(extent, filesystem, root, extent.Strings("extends"), visited); err != nil {
				return err
			}

			if err := k.Load(koanfProvider{extent}, nil, mergeJobsOption); err != nil {
				return err
			}
		}
	}

	return nil
}

func unmarshalConfigs(main, secondary *koanf.Koanf, c *Config) error {
	c.Hooks = make(map[string]*Hook)

	for hookName := range AvailableHooks {
		if !main.Exists(hookName) && !secondary.Exists(hookName) {
			continue
		}

		if err := addHook(hookName, main, secondary, c); err != nil {
			return err
		}
	}

	// For extra non-git hooks.
	// This behavior may be deprecated in next versions.
	// Notice that with append we're allowing extra hooks to be added in local config
	for _, maybeHook := range append(main.Keys(), secondary.Keys()...) {
		if !hookKeyRegexp.MatchString(maybeHook) {
			continue
		}

		matches := hookKeyRegexp.FindStringSubmatch(maybeHook)
		hookName := matches[hookKeyRegexp.SubexpIndex("hookName")]
		if _, ok := c.Hooks[hookName]; ok {
			continue
		}

		if err := addHook(hookName, main, secondary, c); err != nil {
			return err
		}
	}

	// Merge config and unmarshal it
	if err := main.Merge(secondary); err != nil {
		return err
	}

	if err := main.Unmarshal("", c); err != nil {
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

func addHook(name string, main, secondary *koanf.Koanf, c *Config) error {
	mainHook := main.Cut(name)
	overrideHook := secondary.Cut(name)

	// Special merge func to support merging {cmd} templates
	options := koanf.WithMergeFunc(func(src, dest map[string]interface{}) error {
		var destCommands map[string]string

		switch commands := dest["commands"].(type) {
		case map[string]interface{}:
			destCommands = make(map[string]string, len(commands))
			for cmdName, command := range commands {
				switch cmd := command.(type) {
				case map[string]interface{}:
					switch run := cmd["run"].(type) {
					case string:
						destCommands[cmdName] = run
					default:
					}
				default:
				}
			}
		default:
		}

		var destJobs, srcJobs []interface{}
		switch jobs := dest["jobs"].(type) {
		case []interface{}:
			destJobs = jobs
		default:
		}
		switch jobs := src["jobs"].(type) {
		case []interface{}:
			srcJobs = jobs
		default:
		}

		destJobs = mergeJobsSlice(srcJobs, destJobs)

		maps.Merge(src, dest)

		if len(destCommands) > 0 {
			switch commands := dest["commands"].(type) {
			case map[string]interface{}:
				for cmdName, command := range commands {
					switch cmd := command.(type) {
					case map[string]interface{}:
						switch run := cmd["run"].(type) {
						case string:
							newRun := strings.ReplaceAll(run, CMD, destCommands[cmdName])
							command.(map[string]interface{})["run"] = newRun
						default:
						}
					default:
					}
				}
			default:
			}
		}

		if len(destJobs) > 0 {
			dest["jobs"] = destJobs
		}

		return nil
	})

	if err := mainHook.Load(koanfProvider{overrideHook}, nil, options); err != nil {
		return err
	}
	var hook Hook
	if err := mainHook.Unmarshal("", &hook); err != nil {
		return err
	}

	if tags := os.Getenv("LEFTHOOK_EXCLUDE"); tags != "" {
		hook.ExcludeTags = append(hook.ExcludeTags, strings.Split(tags, ",")...)
	}

	c.Hooks[name] = &hook
	return nil
}

// Rewritten from afero.NewIOFS to support opening paths starting with '/'.

type iofs struct {
	fs afero.Fs
}

func newIOFS(filesystem afero.Fs) iofs {
	return iofs{filesystem}
}

func (iofs iofs) Open(name string) (fs.File, error) {
	file, err := iofs.fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open failed: %s: %w", name, err)
	}

	return file, nil
}

type koanfProvider struct {
	k *koanf.Koanf
}

func (k koanfProvider) Read() (map[string]interface{}, error) {
	return k.k.Raw(), nil
}

func (k koanfProvider) ReadBytes() ([]byte, error) {
	panic("not implemented")
}

func mergeJobs(src, dest map[string]interface{}) error {
	srcJobs := make(map[string][]interface{})

	for name, maybeHook := range src {
		switch hook := maybeHook.(type) {
		case map[string]interface{}:
			switch jobs := hook["jobs"].(type) {
			case []interface{}:
				srcJobs[name] = jobs
			default:
			}
		default:
		}
	}

	destJobs := make(map[string][]interface{})
	for name, maybeHook := range dest {
		switch hook := maybeHook.(type) {
		case map[string]interface{}:
			switch jobs := hook["jobs"].(type) {
			case []interface{}:
				destJobs[name] = jobs
			default:
			}
		default:
		}
	}

	if len(srcJobs) == 0 || len(destJobs) == 0 {
		maps.Merge(src, dest)
		return nil
	}

	for hook, newJobs := range srcJobs {
		oldJobs, ok := destJobs[hook]
		if !ok {
			destJobs[hook] = newJobs
			continue
		}

		destJobs[hook] = mergeJobsSlice(newJobs, oldJobs)
	}

	maps.Merge(src, dest)

	for name, maybeHook := range dest {
		if jobs, ok := destJobs[name]; ok {
			switch hook := maybeHook.(type) {
			case map[string]interface{}:
				hook["jobs"] = jobs
			default:
			}
		}
	}

	return nil
}

func mergeJobsSlice(src, dest []interface{}) []interface{} {
	mergeable := make(map[string]map[string]interface{})
	result := make([]interface{}, 0, len(dest))

	for _, maybeJob := range dest {
		switch destJob := maybeJob.(type) {
		case map[string]interface{}:
			switch name := destJob["name"].(type) {
			case string:
				mergeable[name] = destJob
			default:
			}

			result = append(result, maybeJob)
		default:
		}
	}

	for _, maybeJob := range src {
		switch srcJob := maybeJob.(type) {
		case map[string]interface{}:
			switch name := srcJob["name"].(type) {
			case string:
				destJob, ok := mergeable[name]
				if ok {
					var srcSubJobs []interface{}
					var destSubJobs []interface{}

					switch srcGroup := srcJob["group"].(type) {
					case map[string]interface{}:
						switch subJobs := srcGroup["jobs"].(type) {
						case []interface{}:
							srcSubJobs = subJobs
						default:
						}
					default:
					}
					switch destGroup := destJob["group"].(type) {
					case map[string]interface{}:
						switch subJobs := destGroup["jobs"].(type) {
						case []interface{}:
							destSubJobs = subJobs
						default:
						}
					default:
					}

					if len(destSubJobs) != 0 && len(srcSubJobs) != 0 {
						destSubJobs = mergeJobsSlice(srcSubJobs, destSubJobs)
					}

					// Replace possible {cmd} before merging the jobs
					switch srcRun := srcJob["run"].(type) {
					case string:
						switch destRun := destJob["run"].(type) {
						case string:
							newRun := strings.ReplaceAll(srcRun, CMD, destRun)
							srcJob["run"] = newRun
						default:
						}
					default:
					}

					maps.Merge(srcJob, destJob)

					if len(destSubJobs) != 0 {
						switch destGroup := destJob["group"].(type) {
						case map[string]interface{}:
							switch destGroup["jobs"].(type) {
							case []interface{}:
								destGroup["jobs"] = destSubJobs
							default:
							}
						default:
						}
					}
					continue
				}
			default:
			}

			result = append(result, maybeJob)
		default:
		}
	}

	return result
}
