package lefthook

import (
	"fmt"
	"path/filepath"

	"github.com/evilmartians/lefthook/internal/config"
)

const defaultDirMode = 0o755

type AddArgs struct {
	Hook string

	CreateDirs, Force bool
}

func Add(opts *Options, args *AddArgs) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Add(args)
}

// Creates a hook, given in args. The hook is a Lefthook hook.
func (l *Lefthook) Add(args *AddArgs) error {
	if !config.HookAvailable(args.Hook) {
		return fmt.Errorf("Skip adding, hook is unavailable: %s", args.Hook)
	}

	err := l.cleanHook(args.Hook, args.Force || l.Options.Force)
	if err != nil {
		return err
	}

	if err = l.ensureHooksDirExists(); err != nil {
		return err
	}

	err = l.addHook(args.Hook, "")
	if err != nil {
		return err
	}

	if args.CreateDirs {
		global, local := l.getSourceDirs()

		sourceDir := filepath.Join(l.repo.RootPath, global, args.Hook)
		sourceDirLocal := filepath.Join(l.repo.RootPath, local, args.Hook)

		if err = l.Fs.MkdirAll(sourceDir, defaultDirMode); err != nil {
			return err
		}
		if err = l.Fs.MkdirAll(sourceDirLocal, defaultDirMode); err != nil {
			return err
		}
	}

	return nil
}

func (l *Lefthook) getSourceDirs() (global, local string) {
	global = config.DefaultSourceDir
	local = config.DefaultSourceDirLocal

	cfg, err := config.Load(l.Fs, l.repo)
	if err == nil {
		if len(cfg.SourceDir) > 0 {
			global = cfg.SourceDir
		}
		if len(cfg.SourceDirLocal) > 0 {
			local = cfg.SourceDirLocal
		}
	}

	return
}
