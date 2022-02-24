package lefthook

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/log"
)

type UninstallArgs struct {
	KeepConfiguration, Aggressive bool
}

func Uninstall(opts *Options, args *UninstallArgs) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Uninstall(args)
}

func (l *Lefthook) Uninstall(args *UninstallArgs) error {
	if err := l.deleteHooks(args.Aggressive || l.Options.Aggressive); err != nil {
		return err
	}

	if !args.KeepConfiguration {
		for _, glob := range []string{
			"lefthook.y*ml",
			"lefthook-local.y*ml",
		} {
			l.removeFile(filepath.Join(l.repo.RootPath, glob))
		}
	}

	return nil
}

func (l *Lefthook) deleteHooks(force bool) error {
	hooks, err := afero.ReadDir(l.Fs, l.repo.HooksPath)
	if err != nil {
		return err
	}

	for _, file := range hooks {
		hookFile := filepath.Join(l.repo.HooksPath, file.Name())

		// Skip non-lefthook files if removal not forced
		if !l.isLefthookFile(hookFile) && !force {
			continue
		}

		if err := l.Fs.Remove(hookFile); err == nil {
			log.Debugf("%s removed", hookFile)
		} else {
			log.Errorf("Failed removing %s: %s\n", hookFile, err)
		}

		// Recover .old file if exists
		oldHookFile := filepath.Join(l.repo.HooksPath, file.Name()+".old")
		if exists, _ := afero.Exists(l.Fs, oldHookFile); !exists {
			continue
		}

		if err := l.Fs.Rename(oldHookFile, hookFile); err == nil {
			log.Debug(oldHookFile, "renamed to", file.Name())
		} else {
			log.Errorf("Failed renaming %s: %s\n", oldHookFile, err)
		}
	}

	return nil
}

func (l *Lefthook) removeFile(glob string) {
	paths, err := afero.Glob(l.Fs, glob)
	if err != nil {
		log.Errorf("Failed removing configuration files: %s\n", err)
		return
	}

	for _, fileName := range paths {
		if err := l.Fs.Remove(fileName); err == nil {
			log.Debugf("%s removed", fileName)
		} else {
			log.Errorf("Failed removing file %s: %s\n", fileName, err)
		}
	}
}
