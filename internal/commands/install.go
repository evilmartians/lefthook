package commands

import (
	"context"
	"fmt"

	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/evilmartians/lefthook/v2/internal/config"
)

type InstallArgs struct {
	Force          bool
	ResetHooksPath bool
	Hooks          []string
}

func Install(ctx context.Context, app *app.App, args InstallArgs) error {
	configService := app.ConfigService()

	if !configService.Exists() {
		if err := configService.Create(); err != nil {
			return err
		}
	}

	cfg, err := configService.Load()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	var remotesSynced bool
	gitService := app.GitService()
	configService.ForEachRemote(func(remote *config.Remote) {
		if !remote.Configured() {
			return
		}

		if err := gitService.SyncRemote(ctx, remote, args.Force); err != nil {
			app.Logger.Warnf("Couldn't sync from %s. Will continue anyway: %s", remote.GitURL, err)
			return
		}

		remotesSynced = true
	})

	if remotesSynced {
		cfg, err = configService.Reload()
		if err != nil {
			return err
		}
	}

	return app.HooksService().Install(cfg, args.Hooks, args.Force, args.ResetHooksPath)
}
