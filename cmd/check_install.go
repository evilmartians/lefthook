package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func checkInstall() *cli.Command {
	return &cli.Command{
		Name:  "check-install",
		Usage: "check if hooks are installed",
		UsageText: `lefthook check-install – Check if lefthook is installed. Exit codes:
0 – hooks are installed
1 – hooks are not installed or stale`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(false, "auto")
			if err != nil {
				return err
			}

			return l.CheckInstall(ctx)
		},
	}
}
