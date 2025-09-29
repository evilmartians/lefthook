package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/templates"
)

const (
	configFileMode   = 0o666
	checksumFileMode = 0o644
	hooksDirMode     = 0o755
	timestampBase    = 10
	timestampBitsize = 64
)

var (
	lefthookChecksumRegexp = regexp.MustCompile(`(\w+)\s+(\d+)(?:\s+([\w,-]+))?`)
	errNoConfig            = errors.New("no lefthook config found")
)

type InstallArgs struct {
	Force bool
}

func (l *Lefthook) Install(ctx context.Context, args InstallArgs, hooks []string) error {
	cfg, err := l.readOrCreateConfig()
	if err != nil {
		return err
	}

	var remotesSynced bool
	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			if err = l.repo.SyncRemote(remote.GitURL, remote.Ref, args.Force); err != nil {
				log.Warnf("Couldn't sync from %s. Will continue anyway: %s", remote.GitURL, err)
				continue
			}

			remotesSynced = true
		}
	}

	if remotesSynced {
		// Reread the config file with synced remotes
		cfg, err = l.readOrCreateConfig()
		if err != nil {
			return err
		}
	}

	return l.createHooksIfNeeded(cfg, hooks, args.Force)
}

func (l *Lefthook) readOrCreateConfig() (*config.Config, error) {
	log.Debug("config dir: ", l.repo.RootPath)

	if !l.configExists(l.repo.RootPath) {
		log.Info("Config not found, creating...")
		if err := l.createConfig(l.repo.RootPath); err != nil {
			return nil, err
		}
	}

	return config.Load(l.fs, l.repo)
}

func (l *Lefthook) configExists(path string) bool {
	configPath, _ := l.findMainConfig(path)
	return configPath != ""
}

func (l *Lefthook) findMainConfig(path string) (string, error) {
	configOverride := os.Getenv("LEFTHOOK_CONFIG")
	if len(configOverride) != 0 {
		if !filepath.IsAbs(configOverride) {
			configOverride = filepath.Join(path, configOverride)
		}
		if ok, _ := afero.Exists(l.fs, configOverride); !ok {
			return "", fmt.Errorf("couldn't find config from LEFTHOOK_CONFIG: %s", configOverride)
		}
		return configOverride, nil
	}

	for _, name := range config.MainConfigNames {
		for _, extension := range []string{
			".yml", ".yaml", ".toml", ".json",
		} {
			configPath := filepath.Join(path, name+extension)
			if ok, _ := afero.Exists(l.fs, configPath); ok {
				return configPath, nil
			}
		}
	}

	return "", errNoConfig
}

func (l *Lefthook) createConfig(path string) error {
	file := filepath.Join(path, config.DefaultConfigName)

	err := afero.WriteFile(l.fs, file, templates.Config(), configFileMode)
	if err != nil {
		return err
	}

	log.Info("Added config:", file)

	return nil
}

func (l *Lefthook) syncHooks(cfg *config.Config, fetchRemotes bool) (*config.Config, error) {
	var remotesSynced bool
	var err error

	//nolint:nestif
	if fetchRemotes {
		for _, remote := range cfg.Remotes {
			if remote.Configured() && l.shouldRefetch(remote) {
				if err = l.repo.SyncRemote(remote.GitURL, remote.Ref, false); err != nil {
					log.Warnf("Couldn't sync from %s. Will continue anyway: %s", remote.GitURL, err)
					continue
				}

				remotesSynced = true
			}
		}

		if remotesSynced {
			// Reread the config file with synced remotes
			cfg, err = l.readOrCreateConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to reread the config: %w", err)
			}
		}
	}

	ok, hooks := l.checkHooksSynchronized(cfg)
	if ok {
		return cfg, nil
	}

	// Don't rely on config checksum if remotes were refetched
	return cfg, l.createHooksIfNeeded(cfg, hooks, false)
}

func (l *Lefthook) shouldRefetch(remote *config.Remote) bool {
	if remote.Refetch || remote.RefetchFrequency == "always" {
		return true
	}
	if remote.RefetchFrequency == "" || remote.RefetchFrequency == "never" {
		return false
	}

	timedelta, err := time.ParseDuration(remote.RefetchFrequency)
	if err != nil {
		log.Warnf("Couldn't parse refetch frequency %s. Will continue anyway: %s", remote.RefetchFrequency, err)
		return false
	}

	var lastFetchTime time.Time
	remotePath := l.repo.RemoteFolder(remote.GitURL, remote.Ref)
	info, err := l.fs.Stat(filepath.Join(remotePath, ".git", "FETCH_HEAD"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true
		}

		log.Warnf("Failed to detect last fetch time: %s", err)
		return false
	}

	lastFetchTime = info.ModTime()
	return time.Now().After(lastFetchTime.Add(timedelta))
}

func (l *Lefthook) createHooksIfNeeded(cfg *config.Config, hooks []string, force bool) error {
	onlyHooks := make(map[string]struct{})
	for _, hook := range hooks {
		onlyHooks[hook] = struct{}{}
	}

	var success bool
	defer func() {
		if !success {
			log.Info(log.Cyan("sync hooks: ❌"))
		}
	}()

	checksum, err := cfg.Md5()
	if err != nil {
		return fmt.Errorf("could not calculate checksum: %w", err)
	}

	if err = l.ensureHooksDirExists(); err != nil {
		return fmt.Errorf("could not create hooks dir: %w", err)
	}

	rootsMap := make(map[string]struct{})
	for _, hook := range cfg.Hooks {
		for _, command := range hook.Commands {
			if len(command.Root) > 0 {
				root := strings.Trim(command.Root, "/")
				if _, ok := rootsMap[root]; !ok {
					rootsMap[root] = struct{}{}
				}
			}
		}

		collectAllJobRoots(rootsMap, hook.Jobs)
	}
	roots := make([]string, 0, len(rootsMap))
	for root := range rootsMap {
		roots = append(roots, root)
	}

	hookNames := make([]string, 0, len(cfg.Hooks)+1)
	for hook := range cfg.Hooks {
		if _, ok := onlyHooks[hook]; len(onlyHooks) > 0 && !ok {
			log.Debug("skip installing: ", hook)
			continue
		}

		hookNames = append(hookNames, hook)

		if err = l.cleanHook(hook, force); err != nil {
			return fmt.Errorf("could not replace the hook: %w", err)
		}

		templateArgs := templates.Args{
			Rc:                      cfg.Rc,
			AssertLefthookInstalled: cfg.AssertLefthookInstalled,
			Roots:                   roots,
			LefthookExe:             cfg.Lefthook,
		}
		if err = l.addHook(hook, templateArgs); err != nil {
			return fmt.Errorf("could not add the hook: %w", err)
		}
	}

	if len(onlyHooks) == 0 {
		templateArgs := templates.Args{
			Rc:                      cfg.Rc,
			AssertLefthookInstalled: cfg.AssertLefthookInstalled,
			Roots:                   roots,
			LefthookExe:             cfg.Lefthook,
		}
		if err = l.addHook(config.GhostHookName, templateArgs); err != nil {
			return nil
		}
	}

	if err = l.addChecksumFile(checksum, hooks); err != nil {
		return fmt.Errorf("could not create a checksum file: %w", err)
	}

	success = true
	if len(hookNames) > 0 {
		log.Info(log.Cyan("sync hooks: ✔️"), log.Gray("("+strings.Join(hookNames, ", ")+")"))
	} else {
		log.Info(log.Cyan("sync hooks: ✔️ "))
	}

	return nil
}

func collectAllJobRoots(roots map[string]struct{}, jobs []*config.Job) {
	for _, job := range jobs {
		if len(job.Root) > 0 {
			root := strings.Trim(job.Root, "/")
			if _, ok := roots[root]; !ok {
				roots[root] = struct{}{}
			}
		}

		if job.Group != nil {
			collectAllJobRoots(roots, job.Group.Jobs)
		}
	}
}

// checkHooksSynchronized checks is config hooks synchronized and returns the
// list of hooks which are synchronized.
func (l *Lefthook) checkHooksSynchronized(cfg *config.Config) (bool, []string) {
	// Check checksum in a checksum file
	file, err := l.fs.Open(l.checksumFilePath())
	if err != nil {
		return false, nil
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Warnf("Could not close %s: %s", file.Name(), cErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var storedChecksum string
	var storedTimestamp int64
	var storedHooks []string

	// Checksum format:
	// <md5sum> <timestamp> <hook1,hook2,hook3>
	for scanner.Scan() {
		match := lefthookChecksumRegexp.FindStringSubmatch(scanner.Text())
		if match != nil {
			storedChecksum = match[1]
			storedTimestamp, err = strconv.ParseInt(match[2], timestampBase, timestampBitsize)
			if err != nil {
				return false, nil
			}
			if len(match[3]) > 0 {
				storedHooks = strings.Split(match[3], ",")
			}

			break
		}
	}

	if len(storedChecksum) == 0 {
		return false, storedHooks
	}

	configTimestamp, err := l.configLastUpdateTimestamp()
	if err != nil {
		return false, storedHooks
	}

	if storedTimestamp == configTimestamp {
		return true, storedHooks
	}

	configChecksum, err := cfg.Md5()
	if err != nil {
		return false, storedHooks
	}

	return storedChecksum == configChecksum, storedHooks
}

func (l *Lefthook) configLastUpdateTimestamp() (int64, error) {
	configPath, err := l.findMainConfig(l.repo.RootPath)
	if err != nil {
		return 0, err
	}

	config, err := l.fs.Stat(configPath)
	if err != nil {
		return 0, err
	}

	return config.ModTime().Unix(), nil
}

func (l *Lefthook) addChecksumFile(checksum string, hooks []string) error {
	timestamp, err := l.configLastUpdateTimestamp()
	if err != nil {
		return fmt.Errorf("unable to get config update timestamp: %w", err)
	}

	return afero.WriteFile(
		l.fs, l.checksumFilePath(), templates.Checksum(checksum, timestamp, hooks), checksumFileMode,
	)
}

func (l *Lefthook) checksumFilePath() string {
	return filepath.Join(l.repo.InfoPath, config.ChecksumFileName)
}

func (l *Lefthook) ensureHooksDirExists() error {
	exists, err := afero.Exists(l.fs, l.repo.HooksPath)
	if !exists || err != nil {
		err = l.fs.MkdirAll(l.repo.HooksPath, hooksDirMode)
		if err != nil {
			return err
		}
	}

	return nil
}
