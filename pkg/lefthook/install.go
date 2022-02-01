package lefthook

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"regexp"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/config"
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/templates"
)

const (
	CHECKSUM_HOOK_FILENAME = "prepare-commit-msg"
)

type InstallArgs struct {
	Force, Aggressive bool
}

func (l Lefthook) Install(args *InstallArgs) error {
	if err := initRepo(&l); err != nil {
		return err
	}

	cfg, err := l.readOrCreateConfig()
	if err != nil {
		return err
	}

	return l.createHooks(cfg,
		args.Force || args.Aggressive || l.opts.Force || l.opts.Aggressive)
}

func (l Lefthook) readOrCreateConfig() (*config.Config, error) {
	path := l.repo.RootPath()

	log.Debug("Searching config in:", path)

	if !l.configExists(path) {
		log.Info("Config not found, creating...")

		if err := l.createConfig(path); err != nil {
			return nil, err
		}
	}

	return config.Load(l.fs, path)
}

func (l Lefthook) configExists(path string) bool {
	confPath := filepath.Join(path, "lefthook")
	for _, ext := range []string{".yml", ".yaml"} {
		if result, _ := afero.Exists(l.fs, confPath+ext); result {
			return result
		}
	}

	return false
}

func (l Lefthook) createConfig(path string) error {
	file := filepath.Join(path, "lefthook.yml")

	err := afero.WriteFile(l.fs, file, templates.Config(), 0666)
	if err != nil {
		return err
	}

	log.Println("Added config:", file)

	return nil
}

func (l Lefthook) createHooks(cfg *config.Config, force bool) error {
	if force && l.hooksSynchronized() {
		return nil
	}

	configChecksum, err := l.configChecksum()
	if err != nil {
		return err
	}

	gitHooksPath, err := l.repo.HooksPath()
	if err != nil {
		return err
	}

	for hookName, _ := range cfg.Hooks {
		hookPath := filepath.Join(gitHooksPath, hookName)
		if err != nil {
			return err
		}

		err := l.cleanHook(hookName, hookPath, force)
		if err != nil {
			return err
		}

		err = l.addHook(hookName, hookPath, configChecksum)
		if err != nil {
			return err
		}
	}

	l.addHook(
		CHECKSUM_HOOK_FILENAME,
		filepath.Join(gitHooksPath, CHECKSUM_HOOK_FILENAME),
		configChecksum,
	)

	return nil
}

func (l Lefthook) cleanHook(hookName, hookPath string, force bool) error {
	exists, err := afero.Exists(l.fs, hookPath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// Remove lefthook hook
	if l.isLefthookFile(hookPath) {
		if err := l.fs.Remove(hookPath); err != nil {
			return err
		}

		return nil
	}

	// Rename existing user hook
	exists, err = afero.Exists(l.fs, hookPath+".old")
	if err != nil {
		return err
	}
	if exists && !force {
		return errors.New(
			"Can't rename " + hookName + " to " +
				hookName + ".old - file already exists",
		)
	}

	err = l.fs.Rename(hookPath, hookPath+".old")
	if err != nil {
		return err
	}

	log.Println("renamed " + hookPath + " to " + hookPath + ".old")
	return nil
}

func (l Lefthook) addHook(hookName, hookPath, configChecksum string) error {
	err := afero.WriteFile(
		l.fs, hookPath, templates.Hook(hookName, configChecksum), 0755,
	)
	if err != nil {
		return err
	}

	log.Println("Added hook:", hookName)
	return nil
}

func (l Lefthook) hooksSynchronized() bool {
	hooksPath, err := l.repo.HooksPath()
	if err != nil {
		return false
	}

	hookFullPath := filepath.Join(hooksPath, CHECKSUM_HOOK_FILENAME)
	fileContent, err := afero.ReadFile(l.fs, hookFullPath)
	if err != nil {
		return false
	}

	checksum, err := l.configChecksum()
	if err != nil {
		return false
	}

	pattern := regexp.MustCompile(`(?:# lefthook_version: )(\w+)`)
	match := pattern.FindStringSubmatch(string(fileContent))

	return match[1] == checksum
}

func (l Lefthook) configChecksum() (string, error) {
	m, err := afero.Glob(l.fs, filepath.Join(l.repo.RootPath(), "lefthook.*"))
	if err != nil {
		return "", err
	}

	file, err := l.fs.Open(m[0])
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)[:16]), nil
}
