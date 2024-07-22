package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newUpgradeCmd(_ *lefthook.Options) *cobra.Command {
	var yes bool
	upgradeCmd := cobra.Command{
		Use:               "upgrade",
		Short:             "Upgrade lefthook executable",
		Example:           "lefthook upgrade",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return upgrade(yes)
		},
	}

	upgradeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "no prompt")

	return &upgradeCmd
}

func upgrade(ask bool) error {
	return nil
}
