package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

func newDumpCmd(opts *lefthook.Options) *cobra.Command {
	dumpArgs := lefthook.DumpArgs{}
	dumpCmd := cobra.Command{
		Use:               "dump",
		Short:             "Prints config merged from all extensions (in YAML format by default)",
		Example:           "lefthook dump",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(cmd *cobra.Command, args []string) {
			lefthook.Dump(opts, dumpArgs)
		},
	}

	dumpCmd.Flags().BoolVarP(
		&dumpArgs.JSON, "json", "j", false,
		"dump in JSON format",
	)

	dumpCmd.Flags().BoolVarP(
		&dumpArgs.TOML, "toml", "t", false,
		"dump in TOML format",
	)

	return &dumpCmd
}
