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

	if !config.Synchronized() {
		os.Exit(1)
	}

	return nil
}
