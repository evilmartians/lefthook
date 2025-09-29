package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/log"
	ver "github.com/evilmartians/lefthook/internal/version"
)

func version() *cli.Command {
	var verbose bool

	return &cli.Command{
		Name: "version",
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
			log.Println(ver.Version(verbose))
			return nil
		},
	}
}
