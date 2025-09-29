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
		Usage:     "execute a group of hooks",
		UsageText: "lefthook run <hook-name> [args...] [options]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "enable debug logs",
				Destination: &args.Verbose,
			},
			&cli.StringFlag{
				Name:        "colors",
				Usage:       "on, off, or auto (default: auto)",
				Destination: &colors,
			},
			&cli.StringSliceFlag{
				Name:        "job",
				Usage:       "run only jobs with names",
				Destination: &args.RunOnlyJobs,
			},
			&cli.StringSliceFlag{
				Name:        "tag",
				Usage:       "run only jobs with tag names",
				Destination: &args.RunOnlyTags,
			},
			&cli.StringSliceFlag{
				Name:        "command",
				Usage:       "run only commands",
				Destination: &args.RunOnlyCommands,
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Usage:       "exclude files from all templates",
				Destination: &args.Exclude,
			},
			&cli.StringSliceFlag{
				Name:        "file",
				Usage:       "overwrite file templates with files",
				Destination: &args.Files,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "do not skip if no files changed",
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "all-files",
				Usage:       "replace files templates with {all_files}",
				Destination: &args.AllFiles,
			},
			&cli.BoolFlag{
				Name:        "no-auto-install",
				Usage:       "do not implicitly install hooks",
				Destination: &args.NoAutoInstall,
			},
			&cli.BoolFlag{
				Name:        "no-stage-fixed",
				Usage:       "ignore 'stage_fixed: true' setting",
				Destination: &args.NoStageFixed,
			},
			&cli.BoolFlag{
				Name:        "no-tty",
				Usage:       "act as if no TTY is connected",
				Destination: &args.NoTTY,
			},
			&cli.BoolFlag{
				Name:        "skip-lfs",
				Usage:       "do not run LFS hooks",
				Destination: &args.SkipLFS,
			},
			&cli.BoolFlag{
				Name:        "fail-on-changes",
				Usage:       "exit with 1 if some of the files were changed",
				Destination: &args.FailOnChanges,
			},
			&cli.BoolFlag{
				Name:        "files-from-stdin",
				Usage:       "parse filelist from STDIN",
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
