package app

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/templates"
)

const (
	fileMode        = 0o755
	dirMode         = 0o755
	backupPostfix   = ".old"
	lefthookKeyword = "LEFTHOOK"
)

// HooksService works with Git hooks on the file system.
type HooksService struct {
	repo   *git.Repo
	logger *logger.Logger
	config *ConfigService
}

func (s *HooksService) Create(hookName string, args templates.Args) error {
	return afero.WriteFile(
		s.repo.Fs,
		filepath.Join(s.repo.HooksPath, hookName),
		templates.Hook(hookName, args),
		fileMode,
	)
}

func (s *HooksService) Delete(hookName string, force bool) error {
	fs := s.repo.Fs

	hookPath := filepath.Join(s.repo.HooksPath, hookName)
	exists, err := afero.Exists(fs, hookPath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// Just remove lefthook hook
	if s.isLefthookScript(hookPath) {
		return fs.Remove(hookPath)
	}

	// Check if .old file already exists before renaming.
	exists, err = afero.Exists(fs, hookPath+backupPostfix)
	if err != nil {
		return err
	}
	if exists {
		if force {
			s.logger.Infof("File %s.old already exists, overwriting", hookName)
		} else {
			return fmt.Errorf("can't rename %s to %s.old - file already exists", hookName, hookName)
		}
	}

	err = fs.Rename(hookPath, hookPath+backupPostfix)
	if err != nil {
		return err
	}

	s.logger.Infof("Renamed %s to %s.old", hookPath, hookPath)
	return nil
}

// CreateBaseDir ensures Git hooks dir exists.
func (s *HooksService) CreateBaseDir() error {
	fs := s.repo.Fs
	exists, err := afero.Exists(fs, s.repo.HooksPath)
	if !exists || err != nil {
		err = fs.MkdirAll(s.repo.HooksPath, dirMode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *HooksService) isLefthookScript(path string) bool {
	file, err := s.repo.Fs.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			s.logger.Warnf("Could not close %s: %s", file.Name(), cErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), lefthookKeyword) {
			return true
		}
	}
	if err = scanner.Err(); err != nil {
		log.Warnf("Could not read %s: %s", file.Name(), err)
	}

	return false
}

// Install writes Git hook files for the given hooks (or all configured hooks
// if hooks is empty). It first guards against a configured core.hooksPath,
// then writes hook files and a checksum file.
func (s *HooksService) Install(cfg *config.Config, hooks []string, force, resetHooksPath bool) error {
	if err := s.ensureHooksPathUnset(force, resetHooksPath); err != nil {
		return err
	}

	return s.createHooksIfNeeded(cfg, hooks, force)
}

func (s *HooksService) createHooksIfNeeded(cfg *config.Config, hooks []string, force bool) error {
	onlyHooks := make(map[string]struct{})
	for _, hook := range hooks {
		onlyHooks[hook] = struct{}{}
	}

	var success bool
	defer func() {
		if !success {
			s.logger.Info("sync hooks: ❌")
		}
	}()

	checksum, err := cfg.Md5()
	if err != nil {
		return fmt.Errorf("could not calculate checksum: %w", err)
	}

	if err = s.CreateBaseDir(); err != nil {
		return fmt.Errorf("could not create hooks dir: %w", err)
	}

	roots := collectRoots(cfg)

	hookNames := make([]string, 0, len(cfg.Hooks)+1)
	for hook := range cfg.Hooks {
		if _, ok := onlyHooks[hook]; len(onlyHooks) > 0 && !ok {
			s.logger.Debug("skip installing: ", hook)
			continue
		}

		if err = s.Delete(hook, force); err != nil {
			return fmt.Errorf("could not replace the hook: %w", err)
		}

		if _, ok := config.AvailableHooks[hook]; !ok && !cfg.InstallNonGitHooks {
			continue
		}

		hookNames = append(hookNames, hook)

		templateArgs := templates.Args{
			Rc:                      cfg.Rc,
			AssertLefthookInstalled: cfg.AssertLefthookInstalled,
			Roots:                   roots,
			LefthookPath:            cfg.Lefthook,
		}
		if err = s.Create(hook, templateArgs); err != nil {
			return fmt.Errorf("could not add the hook: %w", err)
		}
	}

	if len(onlyHooks) == 0 && len(cfg.Hooks) == 0 {
		templateArgs := templates.Args{
			Rc:                      cfg.Rc,
			AssertLefthookInstalled: cfg.AssertLefthookInstalled,
			Roots:                   roots,
			LefthookPath:            cfg.Lefthook,
		}
		if err = s.Create(config.GhostHookName, templateArgs); err != nil {
			return nil
		}
	}

	if err = s.addChecksumFile(checksum, hooks); err != nil {
		return fmt.Errorf("could not create a checksum file: %w", err)
	}

	success = true
	if len(hookNames) > 0 {
		s.logger.Infof("sync hooks: ✔️ (%s)", strings.Join(hookNames, ", "))
	} else {
		s.logger.Info("sync hooks: ✔️ ")
	}

	return nil
}

func (s *HooksService) addChecksumFile(checksum string, hooks []string) error {
	timestamp, err := s.config.LastUpdated()
	if err != nil {
		return fmt.Errorf("unable to get config update timestamp: %w", err)
	}

	return afero.WriteFile(
		s.repo.Fs,
		s.checksumFilePath(),
		templates.Checksum(checksum, timestamp, hooks),
		checksumFileMode,
	)
}

func (s *HooksService) checksumFilePath() string {
	return filepath.Join(s.repo.InfoPath, config.ChecksumFileName)
}

// ensureHooksPathUnset ensures core.hooksPath is not configured.
//
// In general using lefthook doesn't make sense with global hooks.
// Local hooks make sense only in terms of migration from other hook managers.
func (s *HooksService) ensureHooksPathUnset(force, resetHooksPath bool) error {
	local, global := s.getHooksPathConfig()

	hasLocal := len(local) > 0
	hasGlobal := len(global) > 0

	if !hasLocal && !hasGlobal {
		return nil
	}

	if !force && !resetHooksPath {
		return formatHooksPathError(local, global)
	}

	if hasLocal {
		s.logger.Warnf("core.hooksPath is set locally to '%s'", local)
	}
	if hasGlobal {
		s.logger.Warnf("core.hooksPath is set globally to '%s'", global)
	}

	if resetHooksPath {
		return s.unsetHooksPathConfig(local, global)
	}

	path := local
	if !hasLocal && hasGlobal {
		path = global
	}
	s.logger.Warnf("Installing hooks anyway in '%s'", path)

	return nil
}

func (s *HooksService) getHooksPathConfig() (local, global string) {
	local, _ = s.repo.Git.Cmd([]string{"git", "config", "--local", "core.hooksPath"})
	global, _ = s.repo.Git.Cmd([]string{"git", "config", "--global", "core.hooksPath"})
	return
}

func (s *HooksService) unsetHooksPathConfig(local, global string) error {
	if len(local) > 0 {
		if _, err := s.repo.Git.Cmd([]string{"git", "config", "--local", "--unset-all", "core.hooksPath"}); err != nil {
			return fmt.Errorf("failed to unset local core.hooksPath: %w", err)
		}
		s.logger.Warn("local core.hooksPath has been unset.")
	}

	if len(global) > 0 {
		if _, err := s.repo.Git.Cmd([]string{"git", "config", "--global", "--unset-all", "core.hooksPath"}); err != nil {
			return fmt.Errorf("failed to unset global core.hooksPath: %w", err)
		}
		s.logger.Warn("global core.hooksPath has been unset.")
	}

	return nil
}

func formatHooksPathError(local, global string) error {
	var errMsg strings.Builder
	var hints []string
	hasLocal := len(local) > 0
	hasGlobal := len(global) > 0

	if hasLocal {
		fmt.Fprintf(&errMsg, "core.hooksPath is set locally to '%s'\n", local)
		hints = append(hints, "hint:   git config --unset-all --local core.hooksPath")
	}
	if hasGlobal {
		fmt.Fprintf(&errMsg, "core.hooksPath is set globally to '%s'\n", global)
		hints = append(hints, "hint:   git config --unset-all --global core.hooksPath")
	}
	errMsg.WriteString("\n")
	errMsg.WriteString("hint: Unset it:\n")
	errMsg.WriteString(strings.Join(hints, "\n"))
	errMsg.WriteString("\nhint:\n")
	errMsg.WriteString("hint: Run 'lefthook install --reset-hooks-path' to automatically unset it.\n")

	path := local
	if !hasLocal && hasGlobal {
		path = global
	}
	errMsg.WriteString("hint:\n")
	fmt.Fprintf(&errMsg, "hint: Run 'lefthook install --force' to install hooks anyway in '%s'.", path)

	return errors.New(errMsg.String())
}

func collectRoots(cfg *config.Config) []string {
	rootsMap := make(map[string]struct{})
	for _, hook := range cfg.Hooks {
		for _, command := range hook.Commands {
			if len(command.Root) > 0 {
				rootsMap[strings.Trim(command.Root, "/")] = struct{}{}
			}
		}
		collectAllJobRoots(rootsMap, hook.Jobs)
	}
	roots := make([]string, 0, len(rootsMap))
	for root := range rootsMap {
		roots = append(roots, root)
	}
	return roots
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
