package lefthook

import (
	"os"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

type DumpArgs struct {
	JSON   bool
	TOML   bool
	Format string
}

func Dump(opts *Options, args DumpArgs) {
	lefthook, err := initialize(opts)
	if err != nil {
		log.Errorf("couldn't initialize lefthook: %s\n", err)
		return
	}

	cfg, err := config.Load(lefthook.Fs, lefthook.repo)
	if err != nil {
		log.Errorf("couldn't load config: %s\n", err)
		return
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

	if args.JSON {
		format = config.JSONFormat
	}

	if args.TOML {
		format = config.TOMLFormat
	}

	if err := cfg.Dump(format, os.Stdout); err != nil {
		log.Errorf("couldn't dump config: %s\n", err)
		return
	}
}
