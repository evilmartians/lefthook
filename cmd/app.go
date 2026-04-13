package cmd

import (
	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/spf13/afero"
)

func newApp(verbose bool, colors string) *app.App {
	return app.New(
		git.New(afero.NewOsFs()),
		verbose,
		colors,
	)
}
