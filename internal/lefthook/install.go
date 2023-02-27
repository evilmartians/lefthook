package lefthook

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/templates"
)

const (
	configFileMode   = 0o666
	checksumFileMode = 0o644
	hooksDirMode     = 0o755
	configGlob       = "lefthook.y*ml"
	timestampBase    = 10
	timestampBitsize = 64
)

var lefthookChecksumRegexp = regexp.MustCompile(`(\w+)\s+(\d+)`)

type InstallArgs struct {
	Force, Aggressive bool
}

// Install installs the hooks from config file to the .git/hooks.
func Install(opts *Options, args *InstallArgs) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Install(args)
}

func (l *Lefthook) Install(args *InstallArgs) error {
	cfg, err := l.readOrCreateConfig()
	if err != nil {
		return err
	}

	if cfg.Remote.Configured() {
		if err := l.repo.SyncRemote(cfg.Remote.GitURL, cfg.Remote.Ref); err != nil {
			log.Warnf("Couldn't sync remotes. Will continue without them: %s", err)
		} else {
			// Reread the config file with synced remotes
			cfg, err = l.readOrCreateConfig()
			if err != nil {
				return err
			}
		}
	}

	return l.createHooksIfNeeded(cfg,
		args.Force || args.Aggressive || l.Options.Force || l.Options.Aggressive)
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
	paths, err := afero.Glob(l.Fs, filepath.Join(path, configGlob))
	if err != nil {
		return false
	}

	for _, config := range paths {
		if ok, _ := afero.Exists(l.Fs, config); ok {
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

func (l *Lefthook) createHooksIfNeeded(cfg *config.Config, force bool) error {
	if !force && l.hooksSynchronized() {
		return nil
	}

	log.Info(log.Cyan("SYNCING"))

	checksum, err := l.configChecksum()
	if err != nil {
		return err
	}

	if err = l.ensureHooksDirExists(); err != nil {
		return err
	}

	hookNames := make([]string, 0, len(cfg.Hooks)+1)
	for hook := range cfg.Hooks {
		hookNames = append(hookNames, hook)

		if err = l.cleanHook(hook, force); err != nil {
			return err
		}

		if err = l.addHook(hook, cfg.Rc); err != nil {
			return err
		}
	}

	if err = l.addHook(config.GhostHookName, cfg.Rc); err != nil {
		return nil
	}

	if err = l.addChecksumFile(checksum); err != nil {
		return err
	}

	if len(hookNames) > 0 {
		log.Info(log.Cyan("SERVED HOOKS:"), log.Bold(strings.Join(hookNames, ", ")))
	}

	return nil
}

func (l *Lefthook) hooksSynchronized() bool {
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

	configChecksum, err := l.configChecksum()
	if err != nil {
		return false
	}

	return storedChecksum == configChecksum
}

func (l *Lefthook) configLastUpdateTimestamp() (timestamp int64, err error) {
	m, err := afero.Glob(l.Fs, filepath.Join(l.repo.RootPath, configGlob))
	if err != nil {
		return
	}

	info, err := l.Fs.Stat(m[0])
	if err != nil {
		return
	}

	timestamp = info.ModTime().Unix()
	return
}

func (l *Lefthook) configChecksum() (checksum string, err error) {
	m, err := afero.Glob(l.Fs, filepath.Join(l.repo.RootPath, configGlob))
	if err != nil {
		return
	}

	file, err := l.Fs.Open(m[0])
	if err != nil {
		return
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	checksum = hex.EncodeToString(hash.Sum(nil)[:16])
	return
}

func (l *Lefthook) addChecksumFile(checksum string) error {
	timestamp, err := l.configLastUpdateTimestamp()
	if err != nil {
		return err
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
