package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/evilmartians/lefthook/v2/internal/config"
)

type DumpArgs struct {
	Format string
}

func Dump(ctx context.Context, app *app.App, args DumpArgs) error {
	cfg, err := app.ConfigService().Load()
	if err != nil {
		return err
	}

	var format config.DumpFormat

	switch args.Format {
	case "yaml":
		format = config.YAMLFormat
	case "json":
		format = config.JSONFormat
	case "toml":
		format = config.TOMLFormat
	default:
		format = config.YAMLFormat
	}

	if err := cfg.Dump(format, os.Stdout); err != nil {
		return fmt.Errorf("couldn't dump config: %w", err)
	}

	return nil
}
