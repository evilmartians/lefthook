package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func install() *cli.Command {
	var args command.InstallArgs
	var verbose bool

	return &cli.Command{
		Name:      "install",
		Usage:     "installs Git hook from the configuration",
		UsageText: "lefthook install [hook-names...] [options]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Usage:       "overwrite .old files",
				Aliases:     []string{"f"},
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(verbose, "auto")
			if err != nil {
				return err
			}

			return l.Install(ctx, args, cmd.Args().Slice())
		},
	}
}
