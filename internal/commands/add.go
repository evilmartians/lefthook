package commands

import (
	"context"
	"fmt"

	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/templates"
)

type AddArgs struct {
	HookName string

	CreateDirs, Force bool
}

func Add(ctx context.Context, app *app.App, args AddArgs) error {
	if !config.KnownHook(args.HookName) {
		return fmt.Errorf("skip adding, hook is unavailable: %s", args.HookName)
	}

	hooks := app.HooksService()

	err := hooks.Delete(args.HookName, args.Force)
	if err != nil {
		return err
	}

	if err = hooks.CreateBaseDir(); err != nil {
		return err
	}

	err = hooks.Create(args.HookName, templates.Args{})
	if err != nil {
		return err
	}

	if args.CreateDirs {
		if err = app.MkdirForScripts(args.HookName); err != nil {
			return err
		}
	}

	return nil
}
