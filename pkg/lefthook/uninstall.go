package lefthook

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/log"
)

type UninstallArgs struct {
	KeepConfiguration, Aggressive bool
}

func (l Lefthook) Uninstall(args *UninstallArgs) error {
	if err := initRepo(&l); err != nil {
		return err
	}

	if err := l.deleteHooks(l.opts.Aggressive || args.Aggressive); err != nil {
		return err
	}

	if !args.KeepConfiguration {
		rootPath := l.repo.RootPath()

		for _, glob := range []string{
			"lefthook.y*ml",
			"lefthook-local.y*ml",
		} {
			l.removeFile(filepath.Join(rootPath, glob))
		}
	}

	return nil
}

func (l Lefthook) deleteHooks(force bool) error {
	hooksPath, err := l.repo.HooksPath()
	if err != nil {
		return err
	}

	hooks, err := afero.ReadDir(l.fs, hooksPath)
	if err != nil {
		return err
	}

	for _, file := range hooks {
		hookFile := filepath.Join(hooksPath, file.Name())

		// Skip non-lefthook files if removal not forced
		if !l.isLefthookFile(hookFile) && !force {
			continue
		}

		if err := l.fs.Remove(hookFile); err == nil {
			log.Debug(hookFile, "removed")
		} else {
			log.Errorf("Failed removing %s: %s", hookFile, err)
		}

		// Recover .old file if exists
		oldHookFile := filepath.Join(hooksPath, file.Name()+".old")
		if exists, _ := afero.Exists(l.fs, oldHookFile); !exists {
			continue
		}

		if err := l.fs.Rename(oldHookFile, hookFile); err == nil {
			log.Debug(oldHook, "renamed to", file.Name())
		} else {
			log.Errorf("Failed renaming %s: %s", oldHookFile, err)
		}
	}

	return nil
}

func (l Lefthook) removeFile(glob string) {
	paths, err := afero.Glob(l.fs, glob)
	if err != nil {
		log.Errorf("Failed removing configuration files: %s", err)
		return
	}

	for _, fileName := range paths {
		if err := l.fs.Remove(fileName); err == nil {
			log.Debug(fileName, "removed")
		} else {
			log.Errorf("Failed removing file %s: %s", fileName, err)
		}
	}

	return
}
