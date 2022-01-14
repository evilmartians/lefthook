package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evilmartians/lefthook/pkg/config"
	"github.com/evilmartians/lefthook/pkg/git"
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/templates"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	checksumHookFilename = "prepare-commit-msg"
)

type Install struct {
	*Options

	force      bool
	aggressive bool

	repo *git.Repository
}

func NewInstallCmd(opts *Options) *cobra.Command {
	install := &Install{Options: opts}

	installCmd := cobra.Command{
		Use:   "install",
		Short: "Write basic configuration file in your project repository. Or initialize existed config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install.Run()
		},
	}

	installCmd.Flags().BoolVarP(
		&install.force, "force", "f", false,
		"reinstall hooks without checking config version",
	)
	installCmd.Flags().BoolVarP(
		&install.aggressive, "aggressive", "a", false,
		"remove all hooks from .git/hooks dir and install lefthook hooks",
	)

	return &installCmd
}

func (cmd *Install) Run() error {
	repo, err := git.NewRepository()
	if err != nil {
		return err
	}

	cmd.repo = repo

	cfg, err := cmd.readOrCreateConfig()
	if err != nil {
		return err
	}

	return cmd.createHooks(cfg)
}

func (cmd *Install) readOrCreateConfig() (*config.Config, error) {
	path := cmd.repo.RootPath()
	log.Info("Searching config in:", path)

	if !cmd.configExists(path) {
		log.Info("Config not found, creating...")
		if err := cmd.createConfig(path); err != nil {
			return nil, err
		}
	}

	return config.Load(cmd.fs, path)
}

func (cmd *Install) configExists(path string) bool {
	confPath := filepath.Join(path, "lefthook")
	for _, ext := range []string{".yml", ".yaml"} {
		if result, _ := afero.Exists(cmd.fs, confPath+ext); result {
			return result
		}
	}

	return false
}

func (cmd *Install) createConfig(path string) error {
	file := filepath.Join(path, "lefthook.yml")

	err := afero.WriteFile(cmd.fs, file, templates.Config(), 0666)
	if err != nil {
		return err
	}

	log.Println("Added config:", file)

	return nil
}

func (cmd *Install) createHooks(cfg *config.Config) error {
	if !cmd.Forced() && cmd.hooksSynchronized() {
		return nil
	}

	configChecksum, err := cmd.configChecksum()
	if err != nil {
		return err
	}

	gitHooksPath, err := cmd.repo.HooksPath()
	if err != nil {
		return err
	}
	for hookName, _ := range cfg.Hooks {
		hookPath := filepath.Join(gitHooksPath, hookName)
		if err != nil {
			return err
		}

		err := cmd.cleanHook(hookName, hookPath)
		if err != nil {
			return err
		}

		err = cmd.addHook(hookName, hookPath, configChecksum)
		if err != nil {
			return err
		}
	}

	cmd.addHook(
		checksumHookFilename,
		filepath.Join(gitHooksPath, checksumHookFilename),
		configChecksum,
	)

	return nil
}

func (cmd *Install) cleanHook(hookName, hookPath string) error {
	exists, _ := afero.Exists(cmd.fs, hookPath)
	if !exists {
		return nil
	}

	// Remove lefthook hook
	if cmd.isLefthookHook(hookPath) {
		if err := cmd.fs.Remove(hookPath); err != nil {
			return err
		}

		return nil
	}

	// Rename existing user hook
	exists, _ = afero.Exists(cmd.fs, hookPath+".old")
	if !exists {
		return nil
	}

	if !cmd.Forced() {
		return errors.New(
			"Can't rename " + hookName + " to " +
				hookName + ".old - file already exists",
		)
	}

	err := cmd.fs.Rename(hookPath, hookPath+".old")
	if err != nil {
		return err
	}

	log.Println("renamed " + hookPath + " to " + hookPath + ".old")
	return nil
}

func (cmd *Install) addHook(hookName, hookPath, configChecksum string) error {
	err := afero.WriteFile(
		cmd.fs, hookPath, templates.Hook(hookName, configChecksum), 0755,
	)
	if err != nil {
		return err
	}

	log.Println("Added hook:", hookName)
	return nil
}

func (cmd *Install) Forced() bool {
	return cmd.Options.Force || cmd.Options.Aggressive || cmd.force || cmd.aggressive
}

func (cmd *Install) hooksSynchronized() bool {
	hooksPath, err := cmd.repo.HooksPath()
	if err != nil {
		return false
	}

	hookFullPath := filepath.Join(hooksPath, checksumHookFilename)
	fileContent, err := afero.ReadFile(cmd.fs, hookFullPath)
	if err != nil {
		return false
	}

	checksum, err := cmd.configChecksum()
	if err != nil {
		return false
	}

	pattern := regexp.MustCompile(`(?:# lefthook_version: )(\w+)`)
	match := pattern.FindStringSubmatch(string(fileContent))

	return match[1] == checksum
}

func (cmd *Install) configChecksum() (string, error) {
	m, err := afero.Glob(cmd.fs, filepath.Join(cmd.repo.RootPath(), "lefthook.*"))
	if err != nil {
		return "", err
	}

	file, err := cmd.fs.Open(m[0])
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

func (cmd *Install) isLefthookHook(filePath string) bool {
	file, err := afero.ReadFile(cmd.fs, filePath)
	if err != nil {
		return false
	}

	return strings.Contains(string(file), "LEFTHOOK")
}
