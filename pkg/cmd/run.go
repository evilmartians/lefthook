package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewRunCmd(rootCmd *cobra.Command) {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Execute group of hooks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecutor(args, appFs)
		},
	}

	rootCmd.AddCommand(runCmd)
}

func runExecutor(args []string, appFs afero.Fs) error {
	return nil
}
