package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func uninstall() *cli.Command {
	var args command.UninstallArgs
	var verbose bool

	return &cli.Command{
		Name:  "uninstall",
		Usage: "delete installed hooks",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "remove all Git hooks",
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "remove-configs",
				Usage:       "remove lefthook configs",
				Destination: &args.RemoveConfig,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(verbose, "auto")
			if err != nil {
				return err
			}

			return l.Uninstall(ctx, args)
		},
	}
}
