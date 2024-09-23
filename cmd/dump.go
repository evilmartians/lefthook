package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
)

type dump struct{}

func (dump) New(opts *lefthook.Options) *cobra.Command {
	dumpArgs := lefthook.DumpArgs{}
	dumpCmd := cobra.Command{
		Use:               "dump",
		Short:             "Prints config merged from all extensions (in YAML format by default)",
		Example:           "lefthook dump",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			lefthook.Dump(opts, dumpArgs)
		},
	}

	dumpCmd.Flags().StringVarP(
		&dumpArgs.Format, "format", "f", "yaml", "'yaml', 'toml', or 'json'",
	)

	dumpCmd.Flags().BoolVarP(
		&dumpArgs.JSON, "json", "j", false,
		"dump in JSON format",
	)

	dumpCmd.Flags().BoolVarP(
		&dumpArgs.TOML, "toml", "t", false,
		"dump in TOML format",
	)

	err := dumpCmd.Flags().MarkDeprecated("json", "use --format=json")
	if err != nil {
		log.Warn("Unexpected error:", err)
	}

	err = dumpCmd.Flags().MarkDeprecated("toml", "use --format=toml")
	if err != nil {
		log.Warn("Unexpected error:", err)
	}

	return &dumpCmd
}
