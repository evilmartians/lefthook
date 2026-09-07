package cmd

import (
	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/spf13/afero"
)

func newApp(verbose bool, colors string) (*app.App, error) {
	app, err := app.New(
		afero.NewOsFs(),
		verbose,
		colors,
	)
	if err != nil {
		return nil, err
	}

	return app, nil
}
