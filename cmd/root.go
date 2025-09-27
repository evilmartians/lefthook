package cmd

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/log"
	ver "github.com/evilmartians/lefthook/internal/version"
)

const fullVersionFlag = "full"

func newRootCmd() *cobra.Command {
	options := command.Options{}
	var versionFlag string

	rootCmd := &cobra.Command{
		Use:   "lefthook",
		Short: "CLI tool to manage Git hooks",
		Long: heredoc.Doc(`
				After installation go to your project directory and execute the following command:
				lefthook install
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if versionFlag != "" {
				verbose := versionFlag == fullVersionFlag
				log.Println(ver.Version(verbose))
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided and --version is not set, show help
			if versionFlag == "" {
				_ = cmd.Help()
			}
		},
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

	rootCmd.Flags().StringVarP(
		&versionFlag, "version", "V", "", "show lefthook version (use 'full' for version with commit hash)",
	)
	rootCmd.Flags().Lookup("version").NoOptDefVal = "short"

	for _, subcommand := range commands {
		rootCmd.AddCommand(subcommand.New(&options))
	}

	return rootCmd
}
