package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newRunCmd(opts *lefthook.Options) *cobra.Command {
	runArgs := lefthook.RunArgs{}

	runCmd := cobra.Command{
		Use:     "run hook-name [git args...]",
		Short:   "Execute group of hooks",
		Example: "lefthook run pre-commit",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// args[0] - hook name
			// args[1:] - git hook arguments, number and value depends on the hook
			return lefthook.Run(opts, runArgs, args[0], args[1:])
		},
	}

	runCmd.Flags().BoolVarP(
		&runArgs.NoTTY, "no-tty", "n", false,
		"run hook non-interactively, disable spinner",
	)

	return &runCmd
}
