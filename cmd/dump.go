package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newDumpCmd(opts *lefthook.Options) *cobra.Command {
	dumpCmd := cobra.Command{
		Use:     "dump",
		Short:   "Prints config merged from all extensions",
		Example: "lefthook dump",
		Run: func(cmd *cobra.Command, hooks []string) {
			lefthook.Dump(opts)
		},
	}

	return &dumpCmd
}
