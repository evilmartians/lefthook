package command

import (
	"context"
	"fmt"
	"os"

	"github.com/evilmartians/lefthook/internal/config"
)

type DumpArgs struct {
	Format string
}

func (l *Lefthook) Dump(_ctx context.Context, args DumpArgs) error {
	cfg, err := l.LoadConfig()
	if err != nil {
		return fmt.Errorf("couldn't load config: %w", err)
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
