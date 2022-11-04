package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newUninstallCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.UninstallArgs{}

	uninstallCmd := cobra.Command{
		Use:   "uninstall",
		Short: "Revert install command",
		RunE: func(cmd *cobra.Command, _args []string) error {
			return lefthook.Uninstall(opts, &args)
		},
	}

	uninstallCmd.Flags().BoolVarP(
		&args.Force, "aggressive", "a", false,
		"DEPRECATED: will behave like -f/--force option",
	)

	uninstallCmd.Flags().BoolVarP(
		&args.Force, "force", "f", false,
		"remove all git hooks even not lefthook-related",
	)

	uninstallCmd.Flags().BoolVarP(
		&args.RemoveConfig, "remove-configs", "c", false,
		"remove lefthook.yml and lefthook-local.yml",
	)

	return &uninstallCmd
}
