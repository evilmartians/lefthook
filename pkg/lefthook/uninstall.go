package lefthook

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/log"
)

type UninstallArgs struct {
	KeepConfiguration bool
}

func (l Lefthook) Uninstall(args *UninstallArgs) error {
	if err := initRepo(&l); err != nil {
		return err
	}

	if err := l.deleteHooks(l.opts.Aggressive); err != nil {
		return err
	}

	rootPath := l.repo.RootPath()
	if !args.KeepConfiguration {
		l.removeFile(filepath.Join(rootPath, "lefthook.y*ml"))
		l.removeFile(filepath.Join(rootPath, "lefthook-local.y*ml"))
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
		if l.isLefthookFile(hookFile) || force {
			if err := l.fs.Remove(hookFile); err == nil {
				log.Debug(hookFile, "removed")
			} else {
				log.Errorf("Failed removing %s: %s", hookFile, err)
			}

			oldHook := filepath.Join(hooksPath, file.Name()+".old")
			if exists, _ := afero.Exists(l.fs, oldHook); exists {
				if err := l.fs.Rename(oldHook, hookFile); err == nil {
					log.Debug(oldHook, "renamed to", file.Name())
				} else {
					log.Errorf("Failed renaming %s: %s", oldHook, err)
				}
			}
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
