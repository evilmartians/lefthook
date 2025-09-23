package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/log"
)

func newRootCmd() *cobra.Command {
	options := command.Options{}

	rootCmd := &cobra.Command{
		Use:   "lefthook",
		Short: "CLI tool to manage Git hooks",
		Long: heredoc.Doc(`
				After installation go to your project directory and execute the following command:
				lefthook install
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(
		&options.Verbose, "verbose", "v", false, "verbose output",
	)

	rootCmd.PersistentFlags().StringVar(
		&options.Colors, "colors", "auto", "'auto', 'on', or 'off'",
	)

	rootCmd.PersistentFlags().BoolVar(
		&options.NoColors, "no-colors", false, "disable colored output",
	)

	err := rootCmd.PersistentFlags().MarkDeprecated("no-colors", "use --colors")
	if err != nil {
		log.Warn("Unexpected error:", err)
	}

	for _, subcommand := range commands {
		rootCmd.AddCommand(subcommand.New(&options))
	}

	return rootCmd
}
