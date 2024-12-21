package cmd

import (
	"maps"
	"slices"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
)

type run struct{}

func (run) New(opts *lefthook.Options) *cobra.Command {
	runArgs := lefthook.RunArgs{}

	runHookCompletions := func(cmd *cobra.Command, args []string, toComplete string) (ret []string, compDir cobra.ShellCompDirective) {
		compDir = cobra.ShellCompDirectiveNoFileComp
		if len(args) != 0 {
			return
		}
		ret = lefthook.ConfigHookCompletions(opts)
		other := slices.Sorted(maps.Keys(config.AvailableHooks))
		ret = append(ret, other...)
		return
	}

	runHookCommandCompletions := func(cmd *cobra.Command, args []string, toComplete string) (ret []string, compDir cobra.ShellCompDirective) {
		compDir = cobra.ShellCompDirectiveNoFileComp
		if len(args) == 0 {
			return
		}
		ret = lefthook.ConfigHookCommandCompletions(opts, args[0])
		return
	}

	runCmd := cobra.Command{
		Use:               "run hook-name [git args...]",
		Short:             "Execute group of hooks",
		Example:           "lefthook run pre-commit",
		ValidArgsFunction: runHookCompletions,
		Args:              cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] - hook name
			// args[1:] - git hook arguments, number and value depends on the hook
			return lefthook.Run(opts, runArgs, args[0], args[1:])
		},
	}

	runCmd.Flags().BoolVarP(
		&runArgs.Force, "force", "f", false,
		"force execution of commands that can be skipped",
	)

	runCmd.Flags().BoolVarP(
		&runArgs.NoTTY, "no-tty", "n", false,
		"run hook non-interactively, disable spinner",
	)

	runCmd.Flags().BoolVar(
		&runArgs.AllFiles, "all-files", false,
		"run hooks on all files",
	)

	runCmd.Flags().BoolVar(
		&runArgs.NoAutoInstall, "no-auto-install", false,
		"skip updating git hooks",
	)

	runCmd.Flags().BoolVar(
		&runArgs.SkipLFS, "skip-lfs", false,
		"skip running git lfs",
	)

	runCmd.Flags().BoolVar(
		&runArgs.FilesFromStdin, "files-from-stdin", false,
		"get files from standard input, null-separated",
	)

	runCmd.Flags().StringSliceVar(
		&runArgs.Files, "files", nil,
		"run on specified files, comma-separated",
	)

	runCmd.Flags().StringArrayVar(
		&runArgs.Files, "file", nil,
		"run on specified file (repeat for multiple files). takes precedence over --all-files",
	)

	runCmd.Flags().StringSliceVar(
		&runArgs.RunOnlyCommands, "commands", nil,
		"run only specified commands",
	)

	runCmd.Flags().StringSliceVar(
		&runArgs.RunOnlyJobs, "jobs", nil,
		"run only specified jobs",
	)

	err := runCmd.Flags().MarkDeprecated("files", "use --file flag instead")
	if err != nil {
		log.Warn("Unexpected error:", err)
	}

	_ = runCmd.RegisterFlagCompletionFunc("commands", runHookCommandCompletions)

	return &runCmd
}
