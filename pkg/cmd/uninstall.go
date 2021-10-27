package cmd

import (
	"github.com/spf13/cobra"
)

var keepConfiguration bool

func NewUninstallCmd(opts *Options) *cobra.Command {
	uninstallCmd := cobra.Command{
		Use:   "uninstall",
		Short: "Revert install command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallExecutor(opts)
		},
	}

	uninstallCmd.Flags().BoolVarP(
		&keepConfiguration, "keep-config", "k", false,
		"keep configuration files and source directories present",
	)

	return &uninstallCmd
}

//    deleteHooks
func uninstallExecutor(opts *Options) error {
	//DeleteGitHooks(fs)
	//revertOldGitHooks(fs)
	//if !keepConfiguration {
	//	deleteSourceDirs(fs)
	//	deleteConfig(fs)
	//}
	return nil
}
