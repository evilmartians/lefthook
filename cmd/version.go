package cmd

import "github.com/spf13/cobra"

const version string = "0.7.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show lefthook version",
	Run: func(cmd *cobra.Command, args []string) {
		loggerClient.Info(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
