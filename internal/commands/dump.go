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

	format := parseFormat(args.Format)
	if err := cfg.Dump(format, os.Stdout); err != nil {
		return fmt.Errorf("couldn't dump config: %w", err)
	}

	return nil
}

func parseFormat(arg string) config.DumpFormat {
	switch arg {
	case "yaml":
		return config.YAMLFormat
	case "json":
		return config.JSONFormat
	case "toml":
		return config.TOMLFormat
	default:
		return config.YAMLFormat
	}
}
