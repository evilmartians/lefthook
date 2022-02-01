package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/pkg/lefthook"
)

func NewUninstallCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.UninstallArgs{}

	uninstallCmd := cobra.Command{
		Use:   "uninstall",
		Short: "Revert install command",
		RunE: func(cmd *cobra.Command, _args []string) error {
			return lefthook.Uninstall(opts, &args)
		},
	}

	uninstallCmd.Flags().BoolVarP(
		&args.KeepConfiguration, "keep-config", "k", false,
		"keep configuration files and source directories present",
	)

	return &uninstallCmd
}
