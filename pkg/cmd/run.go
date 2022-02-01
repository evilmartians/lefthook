package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/pkg/lefthook"
)

func NewRunCmd(opts *lefthook.Options) *cobra.Command {
	runCmd := cobra.Command{
		Use:   "run",
		Short: "Execute group of hooks",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return func() error { return nil }()
		},
	}

	return &runCmd
}
