package lefthook

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

type DumpArgs struct {
	JSON bool
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

	if err := cfg.Dump(args.JSON); err != nil {
		log.Errorf("couldn't dump config: %s\n", err)
		return
	}
}
