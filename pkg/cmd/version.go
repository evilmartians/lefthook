package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/pkg/lefthook"
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/version"
)

func NewVersionCmd(opts *lefthook.Options) *cobra.Command {
	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Show lefthook version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println(version.Version())
		},
	}

	return &versionCmd
}
