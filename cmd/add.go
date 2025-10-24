package cmd

import (
	"context"
	_ "embed"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/internal/command"
)

//go:embed add-usage.txt
var addUsageText string

func add() *cli.Command {
	var args command.AddArgs
	var verbose bool

	return &cli.Command{
		Name:      "add",
		Usage:     "add scripts directory and install the hook",
		UsageText: addUsageText,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "create-dirs",
				Aliases:     []string{"dirs"},
				Usage:       "create directories for scripts",
				Destination: &args.CreateDirs,
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

			args.Hook = cmd.Args().Get(0)
			return l.Add(ctx, args)
		},
	}
}
