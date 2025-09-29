package command

import (
	"context"
	"os"

	"github.com/evilmartians/lefthook/internal/config"
)

type installationStatus int

const (
	installed installationStatus = iota
	notInstalled
)

func (l *Lefthook) CheckInstall(_ctx context.Context) error {
	check, err := l.checkInstall()
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

	cfg, err := config.Load(l.fs, l.repo)
	if err != nil {
		return notInstalled, err
	}

	ok, _ := l.checkHooksSynchronized(cfg)
	if !ok {
		return notInstalled, nil
	}

	return installed, nil
}
