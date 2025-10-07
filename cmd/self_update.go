package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/updater"
)

func selfUpdate() *cli.Command {
	var yes, force, verbose bool

	return &cli.Command{
		Name:  "self-update",
		Usage: "update lefthook executable",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "yes",
				Aliases:     []string{"y"},
				Usage:       "do not prompt y/n",
				Destination: &yes,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "force reinstall",
				Destination: &force,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if os.Getenv(command.EnvVerbose) == "1" || os.Getenv(command.EnvVerbose) == "true" {
				verbose = true
			}
			if verbose {
				log.SetLevel(log.DebugLevel)
				log.Debug("Verbose mode enabled")
			}

			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to determine the binary path: %w", err)
			}

			ctxCancel, stop := signal.NotifyContext(ctx, os.Interrupt)
			defer stop()

			return updater.New().SelfUpdate(ctxCancel, updater.Options{
				Yes:     yes,
				Force:   force,
				ExePath: exePath,
			})
		},
	}
}
