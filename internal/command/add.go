package command

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/templates"
)

const defaultDirMode = 0o755

type AddArgs struct {
	Hook string

	CreateDirs, Force bool
}

// Creates a hook, given in args. The hook is a Lefthook hook.
func (l *Lefthook) Add(_ctx context.Context, args AddArgs) error {
	if !config.KnownHook(args.Hook) {
		return fmt.Errorf("skip adding, hook is unavailable: %s", args.Hook)
	}

	err := l.cleanHook(args.Hook, args.Force)
	if err != nil {
		return err
	}

	if err = l.ensureHooksDirExists(); err != nil {
		return err
	}

	err = l.addHook(args.Hook, templates.Args{})
	if err != nil {
		return err
	}

	if args.CreateDirs {
		global, local := l.getSourceDirs()

		sourceDir := filepath.Join(l.repo.RootPath, global, args.Hook)
		sourceDirLocal := filepath.Join(l.repo.RootPath, local, args.Hook)

		if err = l.fs.MkdirAll(sourceDir, defaultDirMode); err != nil {
			return err
		}
		if err = l.fs.MkdirAll(sourceDirLocal, defaultDirMode); err != nil {
			return err
		}
	}

	return nil
}

func (l *Lefthook) getSourceDirs() (global, local string) {
	global = config.DefaultSourceDir
	local = config.DefaultSourceDirLocal

	cfg, err := l.LoadConfig()
	if err == nil {
		if len(cfg.SourceDir) > 0 {
			global = cfg.SourceDir
		}
		if len(cfg.SourceDirLocal) > 0 {
			local = cfg.SourceDirLocal
		}
	}

	return global, local
}
