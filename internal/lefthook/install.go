package lefthook

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gobwas/glob"
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
	lefthookChecksumRegexp = regexp.MustCompile(`(\w+)\s+(\d+)`)
	configGlob             = glob.MustCompile("{.,}lefthook.{yml,yaml,json,toml}")
	errNoConfig            = errors.New("no lefthook config found")
)

// Install installs the hooks from config file to the .git/hooks.
func Install(opts *Options, force bool) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Install(force)
}

func (l *Lefthook) Install(force bool) error {
	cfg, err := l.readOrCreateConfig()
	if err != nil {
		return err
	}

	var remotesSynced bool
	for _, remote := range cfg.Remotes {
		if remote.Configured() {
			if err = l.repo.SyncRemote(remote.GitURL, remote.Ref, force); err != nil {
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

	return l.createHooksIfNeeded(cfg, false, force)
}

func (l *Lefthook) readOrCreateConfig() (*config.Config, error) {
	log.Debug("Searching config in:", l.repo.RootPath)

	if !l.configExists(l.repo.RootPath) {
		log.Info("Config not found, creating...")
		if err := l.createConfig(l.repo.RootPath); err != nil {
			return nil, err
		}
	}

	return config.Load(l.Fs, l.repo)
}

func (l *Lefthook) configExists(path string) bool {
	paths, err := afero.ReadDir(l.Fs, path)
	if err != nil {
		return false
	}

	for _, file := range paths {
		if ok := configGlob.Match(file.Name()); ok {
			return true
		}
	}

	return false
}

func (l *Lefthook) createConfig(path string) error {
	file := filepath.Join(path, config.DefaultConfigName)

	err := afero.WriteFile(l.Fs, file, templates.Config(), configFileMode)
	if err != nil {
		return err
	}

	log.Info("Added config:", file)

	return nil
}

func (l *Lefthook) syncHooks(cfg *config.Config, fetchRemotes bool) (*config.Config, error) {
	var remotesSynced bool
	var err error

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
	}

	if remotesSynced {
		// Reread the config file with synced remotes
		cfg, err = l.readOrCreateConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to reread the config: %w", err)
		}
	}

	// Don't rely on config checksum if remotes were refetched
	return cfg, l.createHooksIfNeeded(cfg, true, false)
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
	info, err := l.Fs.Stat(filepath.Join(remotePath, ".git", "FETCH_HEAD"))
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

func (l *Lefthook) createHooksIfNeeded(cfg *config.Config, checkHashSum, force bool) error {
	if checkHashSum && l.hooksSynchronized(cfg) {
		return nil
	}

	log.Infof("%s", log.Cyan("sync hooks"))

	var success bool
	defer func() {
		if !success {
			log.Info(log.Cyan(": ❌"))
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
	}
	roots := make([]string, 0, len(rootsMap))
	for root := range rootsMap {
		roots = append(roots, root)
	}

	hookNames := make([]string, 0, len(cfg.Hooks)+1)
	for hook := range cfg.Hooks {
		hookNames = append(hookNames, hook)

		if err = l.cleanHook(hook, force); err != nil {
			return fmt.Errorf("could not replace the hook: %w", err)
		}

		templateArgs := templates.Args{
			Rc:                      cfg.Rc,
			AssertLefthookInstalled: cfg.AssertLefthookInstalled,
			Roots:                   roots,
		}
		if err = l.addHook(hook, templateArgs); err != nil {
			return fmt.Errorf("could not add the hook: %w", err)
		}
	}

	templateArgs := templates.Args{
		Rc:                      cfg.Rc,
		AssertLefthookInstalled: cfg.AssertLefthookInstalled,
		Roots:                   roots,
	}
	if err = l.addHook(config.GhostHookName, templateArgs); err != nil {
		return nil
	}

	if err = l.addChecksumFile(checksum); err != nil {
		return fmt.Errorf("could not create a checksum file: %w", err)
	}

	success = true
	if len(hookNames) > 0 {
		log.Info(log.Cyan(": ✔️"), log.Gray("("+strings.Join(hookNames, ", ")+")"))
	} else {
		log.Info(log.Cyan(": ✔️ "))
	}

	return nil
}

func (l *Lefthook) hooksSynchronized(cfg *config.Config) bool {
	// Check checksum in a checksum file
	file, err := l.Fs.Open(l.checksumFilePath())
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var storedChecksum string
	var storedTimestamp int64

	for scanner.Scan() {
		match := lefthookChecksumRegexp.FindStringSubmatch(scanner.Text())
		if len(match) > 1 {
			storedChecksum = match[1]
			storedTimestamp, err = strconv.ParseInt(match[2], timestampBase, timestampBitsize)
			if err != nil {
				return false
			}

			break
		}
	}

	if len(storedChecksum) == 0 {
		return false
	}

	configTimestamp, err := l.configLastUpdateTimestamp()
	if err != nil {
		return false
	}

	if storedTimestamp == configTimestamp {
		return true
	}

	configChecksum, err := cfg.Md5()
	if err != nil {
		return false
	}

	return storedChecksum == configChecksum
}

func (l *Lefthook) configLastUpdateTimestamp() (timestamp int64, err error) {
	paths, err := afero.ReadDir(l.Fs, l.repo.RootPath)
	if err != nil {
		return
	}
	var config os.FileInfo
	for _, file := range paths {
		if ok := configGlob.Match(file.Name()); ok {
			config = file
			break
		}
	}

	if config == nil {
		err = errNoConfig
		return
	}

	timestamp = config.ModTime().Unix()
	return
}

func (l *Lefthook) addChecksumFile(checksum string) error {
	timestamp, err := l.configLastUpdateTimestamp()
	if err != nil {
		return fmt.Errorf("unable to get config update timestamp: %w", err)
	}

	return afero.WriteFile(
		l.Fs, l.checksumFilePath(), templates.Checksum(checksum, timestamp), checksumFileMode,
	)
}

func (l *Lefthook) checksumFilePath() string {
	return filepath.Join(l.repo.InfoPath, config.ChecksumFileName)
}

func (l *Lefthook) ensureHooksDirExists() error {
	exists, err := afero.Exists(l.Fs, l.repo.HooksPath)
	if !exists || err != nil {
		err = l.Fs.MkdirAll(l.repo.HooksPath, hooksDirMode)
		if err != nil {
			return err
		}
	}

	return nil
}
