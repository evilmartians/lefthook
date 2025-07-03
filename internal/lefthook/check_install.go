package lefthook

import (
	"os"

	"github.com/evilmartians/lefthook/internal/config"
)

type installationStatus int

const (
	installed installationStatus = iota
	notInstalled
)

func CheckInstall(opts *Options) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	check, err := lefthook.checkInstall()
	if err != nil {
		return err
	}

	switch check {
	case installed:
		os.Exit(0)
	case notInstalled:
		os.Exit(1)
	}

	return nil
}

func (l *Lefthook) checkInstall() (installationStatus, error) {
	if !l.configExists(l.repo.RootPath) {
		return notInstalled, nil
	}

	cfg, err := config.Load(l.Fs, l.repo)
	if err != nil {
		return 0, err
	}

	if !l.hooksSynchronized(cfg) {
		return notInstalled, nil
	}

	return installed, nil
}
