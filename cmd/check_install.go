package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func checkInstall() *cli.Command {
	return &cli.Command{
		Name: "check-install",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(false, "auto")
			if err != nil {
				return err
			}

			return l.CheckInstall(ctx)
		},
	}
}
