package cmd

import (
	"github.com/spf13/cobra"
)

func NewRunCmd(opts *Options) *cobra.Command {
	runCmd := cobra.Command{
		Use:   "run",
		Short: "Execute group of hooks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecutor(args, opts)
		},
	}

	return &runCmd
}

func runExecutor(args []string, opts *Options) error {
	return nil
}
