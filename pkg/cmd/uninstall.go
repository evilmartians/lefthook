package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var keepConfiguration bool

func NewUninstallCmd(rootCmd *cobra.Command) {
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Revert install command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallExecutor(appFs)
		},
	}

	uninstallCmd.Flags().BoolVarP(&keepConfiguration, "keep-config", "k", false, "keep configuration files and source directories present")

	rootCmd.AddCommand(uninstallCmd)
}

//    deleteHooks
func uninstallExecutor(fs afero.Fs) error {
	//DeleteGitHooks(fs)
	//revertOldGitHooks(fs)
	//if !keepConfiguration {
	//	deleteSourceDirs(fs)
	//	deleteConfig(fs)
	//}
	return nil
}
