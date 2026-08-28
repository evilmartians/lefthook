package cmd

import (
	"fmt"
	"slices"

	"github.com/evilmartians/lefthook/v2/internal/app"
	"github.com/urfave/cli/v3"
)

type autocomplete struct {
	app *app.App
}

func newAutocomplete() *autocomplete {
	app, err := newApp(false, "off")
	if err != nil {
		return &autocomplete{}
	}
	return &autocomplete{
		app: app,
	}
}

func (a *autocomplete) printHookNames() {
	if a.app == nil {
		return
	}

	config := a.app.ConfigService()
	cfg, err := config.Load()
	if err != nil {
		return
	}

	for hook := range cfg.Hooks {
		fmt.Println(hook) //nolint:forbidigo // undecorated stdout is a must
	}
}

func (a *autocomplete) printFlags(cmd *cli.Command) {
	given := cmd.FlagNames()
flags:
	for _, f := range cmd.VisibleFlags() {
		toAdd := make([]string, 0, len(f.Names()))
		for _, fn := range f.Names() {
			// Exclude all aliases of a flag if any of them is already given
			if slices.Contains(given, fn) {
				continue flags
			}
			// Do not bother with single letter flags.
			// If the user knows what they're for, they can just write them (hit the letter instead of tab),
			// no need to clutter the output with them.
			if len(fn) != 1 {
				toAdd = append(toAdd, fn)
			}
		}
		for _, fn := range toAdd {
			fmt.Println("--" + fn) //nolint:forbidigo // undecorated stdout is a must
		}
	}
}
