package app

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/app"
)

type UninstallArgs struct {
	Force        bool
	RemoveConfig bool
}

func Uninstall(ctx context.Context, app *app.App, args UninstallArgs) error {
	return nil
}
