package command

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/config"
)

type UninstallArgs struct {
	Force, RemoveConfig bool
}

func (l *Lefthook) Uninstall(_ctx context.Context, args UninstallArgs) error {
	if err := l.deleteHooks(args.Force); err != nil {
		return err
	}

	err := l.fs.Remove(l.checksumFilePath())
	switch {
	case err == nil:
		l.logger.Debugf("%s removed", l.checksumFilePath())
	case errors.Is(err, os.ErrNotExist):
		l.logger.Debugf("%s not found, skipping", l.checksumFilePath())
	default:
		l.logger.Errorf("Failed removing %s: %s", l.checksumFilePath(), err)
	}

	if args.RemoveConfig {
		for _, name := range append(config.MainConfigNames, config.LocalConfigNames...) {
			for _, extension := range []string{
				".yml", ".yaml", ".toml", ".json",
			} {
				l.removeFile(filepath.Join(l.repo.RootPath, name+extension))
			}
		}
	}

	return l.fs.RemoveAll(l.repo.RemotesFolder())
}

func (l *Lefthook) deleteHooks(force bool) error {
	hooks, err := afero.ReadDir(l.fs, l.repo.HooksPath)
	if err != nil {
		return err
	}

	for _, file := range hooks {
		hookFile := filepath.Join(l.repo.HooksPath, file.Name())

		// Skip non-lefthook files if removal not forced
		if !l.isLefthookFile(hookFile) && !force {
			continue
		}

		if err := l.fs.Remove(hookFile); err == nil {
			l.logger.Debugf("%s removed", hookFile)
		} else {
			l.logger.Errorf("Failed removing %s: %s", hookFile, err)
		}

		// Recover .old file if exists
		oldHookFile := filepath.Join(l.repo.HooksPath, file.Name()+".old")
		if exists, _ := afero.Exists(l.fs, oldHookFile); !exists {
			continue
		}

		if err := l.fs.Rename(oldHookFile, hookFile); err == nil {
			l.logger.Debug(oldHookFile, "renamed to", file.Name())
		} else {
			l.logger.Errorf("Failed renaming %s: %s", oldHookFile, err)
		}
	}

	return nil
}

func (l *Lefthook) removeFile(glob string) {
	paths, err := afero.Glob(l.fs, glob)
	if err != nil {
		l.logger.Errorf("Failed removing configuration files: %s", err)
		return
	}

	for _, fileName := range paths {
		if err := l.fs.Remove(fileName); err == nil {
			l.logger.Debugf("%s removed", fileName)
		} else {
			l.logger.Errorf("Failed removing file %s: %s", fileName, err)
		}
	}
}
