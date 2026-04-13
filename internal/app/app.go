// app is a connecting layer between the CLI commands and actual Lefthook logic.
package app

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
)

type App struct {
	git    *git.Git
	logger *logger.Logger
}

func New(git *git.Git, verbose bool, colors string) *App {
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

	return &App{
		git:    git,
		logger: l,
	}
}

func (app *App) Setup() error {
	return app.git.Setup()
}

func (app *App) Load() (*config.Config, error) {
	if err := app.Setup(); err != nil {
		return nil, err
	}

	cfg, err := config.Load(app.git)

	// Reset loaded colors
	app.setColors(cfg.Colors)

	return cfg, err
}

func (app *App) setColors(colors any) {
	if colors == nil {
		return
	}

	switch colorsTyped := colors.(type) {
	case string:
		switch colorsTyped {
		case "on":
			app.logger.SetColors(logger.DefaultColors)
		case "off":
			app.logger.SetColors(logger.NoColors)
		default:
		}
	case bool:
		if colorsTyped {
			app.logger.SetColors(logger.DefaultColors)
		} else {
			app.logger.SetColors(logger.NoColors)
		}
	case map[string]any:
		app.logger.SetColors(map[logger.Color]color.Color{
			logger.ColorCyan:   lipgloss.Color(colorsTyped["cyan"].(string)),
			logger.ColorGray:   lipgloss.Color(colorsTyped["gray"].(string)),
			logger.ColorGreen:  lipgloss.Color(colorsTyped["green"].(string)),
			logger.ColorRed:    lipgloss.Color(colorsTyped["red"].(string)),
			logger.ColorYellow: lipgloss.Color(colorsTyped["yellow"].(string)),
		})
	default:
	}
}
