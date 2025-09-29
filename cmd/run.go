package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/internal/command"
)

func run() *cli.Command {
	var args command.RunArgs
	var colors string

	return &cli.Command{
		Name:      "run",
		Usage:     "Execute a group of hooks",
		UsageText: "lefthook run <hook-name> [args...] [options]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Destination: &args.Verbose,
			},
			&cli.StringFlag{
				Name:        "colors",
				Usage:       "on, off, or auto (default: auto)",
				Destination: &colors,
			},
			&cli.StringSliceFlag{
				Name:        "jobs",
				Destination: &args.RunOnlyJobs,
			},
			&cli.StringSliceFlag{
				Name:        "tags",
				Destination: &args.RunOnlyTags,
			},
			&cli.StringSliceFlag{
				Name:        "commands",
				Destination: &args.RunOnlyCommands,
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Destination: &args.Exclude,
			},
			&cli.StringSliceFlag{
				Name:        "file",
				Aliases:     []string{"files"},
				Destination: &args.Files,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "all-files",
				Destination: &args.AllFiles,
			},
			&cli.BoolFlag{
				Name:        "no-auto-install",
				Destination: &args.NoAutoInstall,
			},
			&cli.BoolFlag{
				Name:        "no-stage-fixed",
				Destination: &args.NoStageFixed,
			},
			&cli.BoolFlag{
				Name:        "no-tty",
				Destination: &args.NoTTY,
			},
			&cli.BoolFlag{
				Name:        "skip-lfs",
				Destination: &args.SkipLFS,
			},
			&cli.BoolFlag{
				Name:        "fail-on-changes",
				Destination: &args.FailOnChanges,
			},
			&cli.BoolFlag{
				Name:        "files-from-stdin",
				Destination: &args.FilesFromStdin,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			l, err := command.NewLefthook(args.Verbose, colors)
			if err != nil {
				return err
			}

			args.Hook = cmd.Args().Get(0)
			args.GitArgs = cmd.Args().Slice()[1:]
			return l.Run(ctx, args)
		},
	}
}
