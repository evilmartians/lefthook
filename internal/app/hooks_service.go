package app

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/templates"
	"github.com/spf13/afero"
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
