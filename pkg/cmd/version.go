package cmd

import (
	"github.com/evilmartians/lefthook/pkg/log"
	"github.com/evilmartians/lefthook/pkg/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd(opts *Options) *cobra.Command {
	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Show lefthook version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println(version.Version())
		},
	}

	return &versionCmd
}
