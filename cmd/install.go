package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newInstallCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.InstallArgs{}

	installCmd := cobra.Command{
		Use:   "install",
		Short: "Write basic configuration file in your project repository. Or initialize existed config",
		RunE: func(cmd *cobra.Command, _args []string) error {
			return lefthook.Install(opts, &args)
		},
	}

	installCmd.Flags().BoolVarP(
		&args.Force, "force", "f", false,
		"reinstall hooks without checking config version",
	)
	installCmd.Flags().BoolVarP(
		&args.Aggressive, "aggressive", "a", false,
		"remove all hooks from .git/hooks dir and install lefthook hooks",
	)

	return &installCmd
}
