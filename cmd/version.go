package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

func newVersionCmd(opts *lefthook.Options) *cobra.Command {
	var verbose bool

	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Show lefthook version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println(version.Version(verbose))
		},
	}

	versionCmd.Flags().BoolVarP(
		&verbose, "full", "f", false,
		"full version with commit hash",
	)

	return &versionCmd
}
