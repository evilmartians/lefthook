package lefthook

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/config"
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/templates"
)

const (
	checksumHookFilename = "prepare-commit-msg"
	configDefaultName    = "lefthook.yml"
	configGlob           = "lefthook.y*ml"
)

var lefthookChecksumRegexp = regexp.MustCompile(`(?:#\s*lefthook_version:\s+)(\w+)`)

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

	return l.createHooks(cfg,
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

	return config.Load(l.Fs, l.repo.RootPath)
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
	file := filepath.Join(path, configDefaultName)

	err := afero.WriteFile(l.Fs, file, templates.Config(), 0666)
	if err != nil {
		return err
	}

	log.Info("Added config:", file)

	return nil
}

func (l *Lefthook) createHooks(cfg *config.Config, force bool) error {
	if !force && l.hooksSynchronized() {
		return nil
	}

	log.Info(log.Cyan("SYNCING"))

	checksum, err := l.configChecksum()
	if err != nil {
		return err
	}

	hookNames := make([]string, len(cfg.Hooks), len(cfg.Hooks)+1)
	for hook := range cfg.Hooks {
		hookNames = append(hookNames, hook)

		err = l.cleanHook(hook, force)
		if err != nil {
			return err
		}

		err = l.addHook(hook, checksum)
		if err != nil {
			return err
		}
	}

	// Add an informational hook to use for checksum comparation.
	err = l.addHook(checksumHookFilename, checksum)
	if err != nil {
		return err
	}

	hookNames = append(hookNames, checksumHookFilename)
	log.Info(log.Cyan("SERVED HOOKS:"), log.Bold(strings.Join(hookNames, ", ")))

	return nil
}

func (l *Lefthook) hooksSynchronized() bool {
	checksum, err := l.configChecksum()
	if err != nil {
		return false
	}

	// Check checksum in a checksum file
	hookFullPath := filepath.Join(l.repo.HooksPath, checksumHookFilename)
	file, err := l.Fs.Open(hookFullPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		match := lefthookChecksumRegexp.FindStringSubmatch(scanner.Text())
		if len(match) > 1 && match[1] == checksum {
			return true
		}
	}

	return false
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
