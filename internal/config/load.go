package config

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/parsers/yaml"
	kfs "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
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
	Extensions       = []string{
		".yml",
		".yaml",
		".json",
		".jsonc",
		".toml",
	}
	parsers = map[string]koanf.Parser{
		".yml":   yaml.Parser(),
		".yaml":  yaml.Parser(),
		".json":  json.Parser(),
		".jsonc": jsoncParser(),
		".toml":  toml.Parser(),
	}

	mergeJobsOption = koanf.WithMergeFunc(mergeHooks)
)

// ConfigNotFoundError.
type ConfigNotFoundError struct {
	message string
}

func (err ConfigNotFoundError) Error() string {
	return err.message
}

type Loader struct {
	repo   *git.Repo
	logger *logger.Logger
}

func NewLoader(repo *git.Repo, logger *logger.Logger) *Loader {
	return &Loader{
		repo:   repo,
		logger: logger,
	}
}

// loadConfig loads the config at the given path.
func (l *Loader) loadConfig(k *koanf.Koanf, path string) error {
	extension := filepath.Ext(path)
	l.logger.Debug("loading config: ", path)
	if err := k.Load(kfs.Provider(newIOFS(l.repo.Fs), path), parsers[extension], mergeJobsOption); err != nil {
		return err
	}

	return nil
}

// loadFirst loads the first existing config from given names and supported extensions.
func (l *Loader) loadFirst(k *koanf.Koanf, root string, names []string) error {
	for _, extension := range Extensions {
		for _, name := range names {
			configPath := filepath.Join(root, name+extension)
			if ok, _ := afero.Exists(l.repo.Fs, configPath); !ok {
				continue
			}

			return l.loadConfig(k, configPath)
		}
	}

	return ConfigNotFoundError{fmt.Sprintf("No config files with names %q have been found in \"%s\"", names, root)}
}

// loadFirstMain loads the main config (e.g. lefthook.yml) or fallbacks to local config (e.g. lefthook-local.yml).
func loadFirstMain(k *koanf.Koanf, filesystem afero.Fs, root string) error {
	err := loadFirst(k, filesystem, root, MainConfigNames)
	if ok := errors.As(err, &ConfigNotFoundError{}); ok {
		var hasLocalConfig bool
	OUT:
		for _, extension := range Extensions {
			for _, name := range LocalConfigNames {
				if ok, _ := afero.Exists(filesystem, filepath.Join(root, name+extension)); ok {
					hasLocalConfig = true
					break OUT
				}
			}
		}
		if !hasLocalConfig {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (l *Loader) loadMain(root string) (*koanf.Koanf, error) {
	main := koanf.New(".")

	configOverridePath := os.Getenv("LEFTHOOK_CONFIG")
	if len(configOverridePath) == 0 {
		if err := loadFirstMain(main, l.repo.Fs, root); err != nil {
			return nil, err
		}

		return main, nil
	}

	if !filepath.IsAbs(configOverridePath) {
		configOverride = filepath.Join(root, configOverridePath)
	}
	if ok, _ := afero.Exists(l.repo.Fs, configOverride); !ok {
		return nil, ConfigNotFoundError{fmt.Sprintf("Config file \"%s\" not found!", configOverride)}
	}

	if err := l.loadConfig(main, configOverridePath); err != nil {
		return nil, err
	}

	return main, nil
}

func (l *Loader) LoadSecondary(main *koanf.Koanf) (*koanf.Koanf, error) {
	// Save `extends` and `remotes`
	extends := main.Strings("extends")
	var remotes []*Remote
	if err := main.Unmarshal("remotes", &remotes); err != nil {
		return nil, err
	}

	secondary := koanf.New(".")

	// Load main `extends`
	if err := extend(secondary, l.repo.Fs, l.repo.RootPath, extends); err != nil {
		return nil, err
	}

	// Some extends required the other extends and changed the list
	// We don't want to load those extends again, so unsetting them.
	if !slices.Equal(secondary.Strings("extends"), extends) {
		secondary.Delete("extends")
	}

	// Load main `remotes`
	if err := l.loadRemotes(secondary, remotes); err != nil {
		return nil, err
	}

	// Don't allow to set `lefthook` field from a remote config
	secondary.Delete("lefthook")

	// Load optional local config (e.g. lefthook-local.yml)
	var noLocal bool
	if err := loadFirst(secondary, l.repo.Fs, l.repo.RootPath, LocalConfigNames); err != nil {
		if ok := errors.As(err, &ConfigNotFoundError{}); !ok {
			return nil, err
		}
		noLocal = true
	}

	// Load local `extends`
	localExtends := secondary.Strings("extends")
	if !noLocal && !slices.Equal(extends, localExtends) {
		if err := extend(secondary, l.repo.Fs, l.repo.RootPath, localExtends); err != nil {
			return nil, err
		}
	}

	return secondary, nil
}

func (l *Loader) LoadKoanf() (*koanf.Koanf, *koanf.Koanf, error) {
	// Load main lefthook.yml
	main, err := loadMain(l.repo.Fs, l.repo.RootPath)
	if err != nil {
		return nil, nil, err
	}

	// Load secondary extends, remotes and lefthook-local.yml
	secondary, err := l.LoadSecondary(main)
	if err != nil {
		return nil, nil, err
	}

	return main, secondary, nil
}

// Load loads configs from the given directory with extensions.
func (l *Loader) Load() (*Config, error) {
	main, secondary, err := l.LoadKoanf()
	if err != nil {
		return nil, err
	}

	return l.Unmarshal(main, secondary)
}

func (l *Loader) Unmarshal(main *koanf.Koanf, secondary *koanf.Koanf) (*Config, error) {
	var config Config

	config.SourceDir = DefaultSourceDir
	config.SourceDirLocal = DefaultSourceDirLocal

	if err := unmarshalConfigs(main, secondary, &config); err != nil {
		return nil, err
	}

	switch colors := config.Colors.(type) {
	case string:
		switch colors {
		case "on":
			l.logger.SetColors(logger.DefaultColors)
		case "off":
			l.logger.SetColors(logger.NoColors)
		}
	case bool:
		if colors {
			l.logger.SetColors(logger.DefaultColors)
		} else {
			l.logger.SetColors(logger.NoColors)
		}
	case map[string]any:
		newColors := make(map[logger.Color]color.Color)
		maps.Copy(newColors, logger.DefaultColors)

		for name, code := range colors {
			var colorCode string
			switch cCode := code.(type) {
			case int:
				colorCode = strconv.Itoa(cCode)
			case string:
				colorCode = code
			default:
				continue
			}

			switch name {
			case "cyan":
				newColors[logger.ColorCyan] = lipgloss.Color(colorCode)
			case "gray":
				newColors[logger.ColorGray] = lipgloss.Color(colorCode)
			case "green":
				newColors[logger.ColorGreen] = lipgloss.Color(colorCode)
			case "red":
				newColors[logger.ColorRed] = lipgloss.Color(colorCode)
			case "yellow":
				newColors[logger.ColorYellow] = lipgloss.Color(colorCode)
			}
		}
		l.logger.SetColors(newColors)
	}

	return &config, nil
}

// loadRemotes merges remote configs to the current one.
func (l *Loader) loadRemotes(k *koanf.Koanf, remotes []*Remote) error {
	for _, remote := range remotes {
		if !remote.Configured() {
			continue
		}

		if len(remote.Configs) == 0 {
			remote.Configs = append(remote.Configs, DefaultConfigName)
		}

		for _, config := range remote.Configs {
			remotePath := l.repo.RemoteFolder(remote.GitURL, remote.Ref)
			configFile := config
			configPath := filepath.Join(remotePath, configFile)

			l.logger.Debugf("Merging remote config: %s: %s", remote.GitURL, configPath)

			if ok, err := afero.Exists(l.repo.Fs, configPath); !ok || err != nil {
				continue
			}

			parser, ok := parsers[filepath.Ext(configPath)]
			if !ok {
				return fmt.Errorf("can't parse config '%[1]s', file has unsupported or no extension\nhint: rename %[1]s to %[1]s.yml", configPath)
			}

			if err := k.Load(kfs.Provider(newIOFS(l.repo.Fs), configPath), parser, mergeJobsOption); err != nil {
				return err
			}

			extends := k.Strings("extends")
			if err := extend(k, l.repo.Fs, filepath.Dir(configPath), extends); err != nil {
				return err
			}
		}

		// Reset extends to omit issues when extending with remote extends.
		if err := k.Set("extends", []string(nil)); err != nil {
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
	// Notice that with append we're allowing extra hooks to be added in local config
	for _, maybeHook := range append(main.Keys(), secondary.Keys()...) {
		matches := hookKeyRegexp.FindStringSubmatch(maybeHook)
		if matches == nil {
			continue
		}
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

	return nil
}

func addHook(name string, main, secondary *koanf.Koanf, c *Config) error {
	mainHook := main.Cut(name)
	overrideHook := secondary.Cut(name)

	// Special merge func to support merging {cmd} templates
	options := koanf.WithMergeFunc(func(src, dest map[string]any) error {
		var destCommands map[string]string

		switch commands := dest["commands"].(type) {
		case map[string]any:
			destCommands = make(map[string]string, len(commands))
			for cmdName, command := range commands {
				switch cmd := command.(type) {
				case map[string]any:
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

		destJobs := anySlice(dest, "jobs")
		srcJobs := anySlice(src, "jobs")
		destSetup := anySlice(dest, "setup")
		srcSetup := anySlice(src, "setup")

		destJobs = mergeJobsSlice(srcJobs, destJobs)
		destSetup = slices.Concat(srcSetup, destSetup)

		maps.Merge(src, dest)

		if len(destCommands) > 0 {
			switch commands := dest["commands"].(type) {
			case map[string]any:
				for cmdName, command := range commands {
					switch cmd := command.(type) {
					case map[string]any:
						switch run := cmd["run"].(type) {
						case string:
							newRun := strings.ReplaceAll(run, CMD, destCommands[cmdName])
							command.(map[string]any)["run"] = newRun
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
		if len(destSetup) > 0 {
			dest["setup"] = destSetup
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
	// Assign custom hook name
	hook.Name = name

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

func (k koanfProvider) Read() (map[string]any, error) {
	return k.k.Raw(), nil
}

func (k koanfProvider) ReadBytes() ([]byte, error) {
	panic("not implemented")
}

// mergeHooks merges `jobs` and `setup` settings.
//
// `jobs` settings get overwritten by name or get appended to the end.
// `setup` always get prepended.
func mergeHooks(src, dest map[string]any) error {
	srcJobs := extractHookSlices(src, "jobs")
	destJobs := extractHookSlices(dest, "jobs")
	srcSetup := extractHookSlices(src, "setup")
	destSetup := extractHookSlices(dest, "setup")

	if (len(srcJobs) == 0 || len(destJobs) == 0) && (len(srcSetup) == 0 || len(destSetup) == 0) {
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

	for hook, newSetup := range srcSetup {
		oldSetup, ok := destSetup[hook]
		if !ok {
			destSetup[hook] = newSetup
			continue
		}

		destSetup[hook] = slices.Concat(oldSetup, newSetup)
	}

	maps.Merge(src, dest)

	for name, maybeHook := range dest {
		if jobs, ok := destJobs[name]; ok {
			switch hook := maybeHook.(type) {
			case map[string]any:
				hook["jobs"] = jobs
			default:
			}
		}
	}

	for name, maybeHook := range dest {
		if setup, ok := destSetup[name]; ok {
			switch hook := maybeHook.(type) {
			case map[string]any:
				hook["setup"] = setup
			default:
			}
		}
	}

	return nil
}

// anySlice extracts a []any value from a string-keyed map, returning nil if absent or wrong type.
func anySlice(m map[string]any, key string) []any {
	v, _ := m[key].([]any)
	return v
}

// extractHookSlices builds a name→[]any map from a hooks map for the given field key.
func extractHookSlices(hooks map[string]any, key string) map[string][]any {
	result := make(map[string][]any)
	for name, maybeHook := range hooks {
		if hook, ok := maybeHook.(map[string]any); ok {
			if v, ok := hook[key].([]any); ok {
				result[name] = v
			}
		}
	}
	return result
}

func mergeJobsSlice(src, dest []any) []any {
	mergeable := make(map[string]map[string]any)
	result := make([]any, 0, len(dest))

	// Pass 1: index dest jobs by name and preserve order.
	for _, maybeJob := range dest {
		switch destJob := maybeJob.(type) {
		case map[string]any:
			switch name := destJob["name"].(type) {
			case string:
				mergeable[name] = destJob
			default:
			}

			result = append(result, maybeJob)
		default:
		}
	}

	// Pass 2: merge src jobs into dest by name, or append unnamed ones.
	for _, maybeJob := range src {
		switch srcJob := maybeJob.(type) {
		case map[string]any:
			switch name := srcJob["name"].(type) {
			case string:
				destJob, ok := mergeable[name]
				if ok {
					var srcSubJobs []any
					var destSubJobs []any

					switch srcGroup := srcJob["group"].(type) {
					case map[string]any:
						switch subJobs := srcGroup["jobs"].(type) {
						case []any:
							srcSubJobs = subJobs
						default:
						}
					default:
					}
					switch destGroup := destJob["group"].(type) {
					case map[string]any:
						switch subJobs := destGroup["jobs"].(type) {
						case []any:
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
						case map[string]any:
							switch destGroup["jobs"].(type) {
							case []any:
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
