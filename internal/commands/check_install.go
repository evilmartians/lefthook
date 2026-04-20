package commands

import (
	"context"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/app"
)

func CheckInstall(ctx context.Context, app *app.App) error {

	config := app.ConfigService()

	if !config.Exists() {
		os.Exit(1)
	}

	// check, err := l.checkInstall()
	// if err != nil {
	// 	return err
	// }

	// switch check {
	// case installed:
	// 	os.Exit(0)
	// case notInstalled:
	// 	os.Exit(1)
	// }

	// return nil
}

func (l *Lefthook) checkInstall() (installationStatus, error) {
	cfg, err := l.LoadConfig()
	if err != nil {
		return notInstalled, err
	}

	ok, _ := l.checkHooksSynchronized(cfg)
	if !ok {
		return notInstalled, nil
	}

	return installed, nil
}
