package lefthook

import (
	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

func Dump(opts *Options) {
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

	if err := cfg.Dump(); err != nil {
		log.Errorf("couldn't dump config: %s\n", err)
		return
	}
}
