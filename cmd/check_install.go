package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

type checkInstall struct{}

func (checkInstall) New(opts *lefthook.Options) *cobra.Command {
	checkInstallCmd := cobra.Command{
		Use:   "check-install",
		Short: "Check if lefthook is installed. Return codes: 0 - installed, 1 - not installed / needs reinstall.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _args []string) error {
			return lefthook.CheckInstall(opts)
		},
	}

	return &checkInstallCmd
}
