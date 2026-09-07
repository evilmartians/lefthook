package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/internal/commands"
)

func install() *cli.Command {
	var args commands.InstallArgs
	var verbose bool

	return &cli.Command{
		Name:      "install",
		Usage:     "install Git hook from the config or create a blank lefthook.yml",
		UsageText: "lefthook install [hook-names...] [options]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Usage:       "overwrite .old files and proceed even if core.hooksPath is set",
				Aliases:     []string{"f"},
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "reset-hooks-path",
				Usage:       "automatically unset core.hooksPath configuration",
				Aliases:     []string{"r"},
				Destination: &args.ResetHooksPath,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			app, err := newApp(verbose, "")
			if err != nil {
				return err
			}
			args.Hooks = cmd.Args().Slice()

			return commands.Install(ctx, app, args)
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			autocomplete := newAutocomplete()
			autocomplete.printFlags(cmd)
			autocomplete.printHookNames()
		},
	}
}
