// app is a connecting layer between the CLI commands and actual Lefthook logic.
package app

import (
	"os"
	"path/filepath"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/spf13/afero"
)

type App struct {
	repo   *git.Repo
	logger *logger.Logger

	config *ConfigService
	hooks  *HooksService
}

func New(fs afero.Fs, verbose bool, colors string) (*App, error) {
	l := logger.New(os.Stdout)
	switch colors {
	case "on", "yes", "true", "1":
		l.SetColors(logger.DefaultColors)
	case "off", "no", "false", "0":
		l.SetColors(logger.NoColors)
	}

	if verbose {
		l.SetLevel(logger.LevelDebug)
	}

	repo, err := git.NewRepo(fs, l)
	if err != nil {
		return nil, err
	}

	return &App{
		repo:   repo,
		logger: l,
	}, nil
}

// MkdirForScripts creates dirs in configured folders to auto-load user scripts by hook and script names.
func (app *App) MkdirForScripts(hookName string) error {
	sourceDirs, err := app.ConfigService().SourceDirs()
	if err != nil {
		return err
	}

	for _, sourceDir := range sourceDirs {
		sourceDir = filepath.Join(app.repo.RootPath, sourceDir, hookName)

		if err := app.repo.Fs.MkdirAll(sourceDir, dirMode); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) ConfigService() *ConfigService {
	if app.config != nil {
		return app.config
	}

	app.config = &ConfigService{
		repo:   app.repo,
		logger: app.logger,
	}

	return app.config
}

func (app *App) HooksService() *HooksService {
	if app.hooks != nil {
		return app.hooks
	}

	app.hooks = &HooksService{
		repo:   app.repo,
		logger: app.logger,
	}

	return app.hooks
}
