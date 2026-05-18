package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/internal/command"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	ver "github.com/evilmartians/lefthook/v2/internal/version"
)

func version() *cli.Command {
	var verbose bool

	return &cli.Command{
		Name:  "version",
		Usage: "print version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "full",
				Aliases:     []string{"f"},
				Destination: &verbose,
			},
		},
		Action: func(_ctx context.Context, cmd *cli.Command) error {
			logger.New(os.Stdout).Info(ver.Version(verbose))

			return nil
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			command.ShellCompleteFlags(cmd)
		},
	}
}
