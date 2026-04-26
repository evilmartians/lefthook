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

	if _, err := configService.Load(); err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	var ok bool
	gitService := app.GitService()
	configService.ForEachRemote(func(remote *config.Remote) {
		if !remote.Configured() {
			return
		}

		if err := gitService.SyncRemote(ctx, remote, args.Force); err != nil {
			app.Logger.Warnf("Couldn't sync from %s. Will continue anyway: %s", remote.GitURL, err)
			return
		}

		ok = true
	})
	if ok {
		if _, err := configService.Reload(); err != nil {
			return err
		}
	}

	return app.HooksService().Install(cfg, args.Hooks, args.Force)

	// var remotesSynced bool
	// for _, remote := range cfg.Remotes {
	// 	if remote.Configured() {
	// 		if err = l.repo.SyncRemote(remote.GitURL, remote.Ref, args.Force); err != nil {
	// 			log.Warnf("Couldn't sync from %s. Will continue anyway: %s", remote.GitURL, err)
	// 			continue
	// 		}

	// 		remotesSynced = true
	// 	}
	// }

	// if remotesSynced {
	// 	// Reread the config file with synced remotes
	// 	cfg, err = l.readOrCreateConfig()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// return app.HooksService.Install(cfg, hooks, args)
}
