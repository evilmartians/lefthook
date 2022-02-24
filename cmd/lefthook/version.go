package main

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

func newVersionCmd(opts *lefthook.Options) *cobra.Command {
	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Show lefthook version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println(version.Version())
		},
	}

	return &versionCmd
}
