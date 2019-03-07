package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const (
	version string = "0.2.0"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show hookah version",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
