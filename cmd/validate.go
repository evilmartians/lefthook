package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func validate() *cli.Command {
	var args command.ValidateArgs
	var verbose bool

	return &cli.Command{
		Name:  "validate",
		Usage: "validate lefthook config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(verbose, "auto")
			if err != nil {
				return nil
			}

			return l.Validate(ctx, args)
		},
	}
}
