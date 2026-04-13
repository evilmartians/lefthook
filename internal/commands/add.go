package commands

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/app"
)

type AddArgs struct {
	Hook string

	CreateDirs, Force bool
}

func Add(ctx context.Context, app *app.App, args AddArgs) error {
	return nil
}
