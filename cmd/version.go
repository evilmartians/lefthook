package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/log"
	ver "github.com/evilmartians/lefthook/internal/version"
)

type version struct{}

func (version) New(_opts *command.Options) *cobra.Command {
	var verbose bool

	versionCmd := cobra.Command{
		Use:               "version",
		Short:             "Show lefthook version",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println(ver.Version(verbose))
		},
	}

	versionCmd.Flags().BoolVarP(
		&verbose, "full", "f", false,
		"full version with commit hash",
	)

	return &versionCmd
}
