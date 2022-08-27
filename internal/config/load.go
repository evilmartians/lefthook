package config

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultSourceDir      = ".lefthook"
	DefaultSourceDirLocal = ".lefthook-local"
	DefaultColorsEnabled  = true
)

var hookKeyRegexp = regexp.MustCompile(`^(?P<hookName>[^.]+)\.(scripts|commands)`)

// Loads configs from the given directory with extensions.
func Load(fs afero.Fs, path string) (*Config, error) {
	global, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	extends, err := mergeAll(fs, path)
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
func mergeAll(fs afero.Fs, path string) (*viper.Viper, error) {
	extends, err := read(fs, path, "lefthook")
	if err != nil {
		return nil, err
	}

	if err := remotes(fs, extends); err != nil {
		return nil, err
	}

	if err := extend(fs, extends); err != nil {
		return nil, err
	}

	extends.SetConfigName("lefthook-local")
	if err := extends.MergeInConfig(); err != nil {
		if _, notFoundErr := err.(viper.ConfigFileNotFoundError); !notFoundErr {
			return nil, err
		}
	}

	if err := remotes(fs, extends); err != nil {
		return nil, err
	}

	if err := extend(fs, extends); err != nil {
		return nil, err
	}

	return extends, nil
}

func remotes(fs afero.Fs, v *viper.Viper) error {
	var remotes []Remote
	err := v.UnmarshalKey("remotes", &remotes)
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		return nil
	}

	var (
		wg           sync.WaitGroup
		eg           errgroup.Group
		configPaths  []string
		configPathCh = make(chan string)
	)

	wg.Add(1)
	go func() {
		for configPath := range configPathCh {
			configPaths = append(configPaths, configPath)
		}
		wg.Done()
	}()

	for i := range remotes {
		remote := remotes[i]
		eg.Go(func() error {
			dir, err := git.InitRemote(fs, remote.URL, remote.Rev)
			if err != nil {
				return err
			}

			for _, path := range remote.Configs {
				configPathCh <- filepath.Join(dir, path)
			}
			return nil
		})
	}

	// Wait on errgroup to finish before closing the channel.
	err = eg.Wait()
	close(configPathCh)
	if err != nil {
		return err
	}

	// Wait for all of the configPaths to be added.
	wg.Wait()

	// Stable sort to ensure that the merge order is deterministic.
	sort.SliceStable(configPaths, func(i, j int) bool { return configPaths[i] < configPaths[j] })

	for _, configPath := range configPaths {
		log.Infof("Merging remote path: %v\n", configPath)
		if err := merge(fs, configPath, v); err != nil {
			return err
		}
	}
	return nil
}

func extend(fs afero.Fs, v *viper.Viper) error {
	for _, path := range v.GetStringSlice("extends") {
		if err := merge(fs, path, v); err != nil {
			return err
		}
	}
	return nil
}

func merge(fs afero.Fs, path string, v *viper.Viper) error {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	another, err := read(fs, filepath.Dir(path), name)
	if err != nil {
		return err
	}
	if err = v.MergeConfigMap(another.AllSettings()); err != nil {
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
	// This behavior will be deprecated in next versions.
	for _, maybeHook := range base.AllKeys() {
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

	resultHook.processDeprecations()

	c.Hooks[hookName] = resultHook

	return nil
}
