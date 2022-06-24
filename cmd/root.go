package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

var commands = [...]func(*lefthook.Options) *cobra.Command{
	newVersionCmd,
	newAddCmd,
	newInstallCmd,
	newUninstallCmd,
	newRunCmd,
}

func newRootCmd() *cobra.Command {
	options := lefthook.Options{
		Fs: afero.NewOsFs(),
	}

	rootCmd := &cobra.Command{
		Use:   "lefthook",
		Short: "CLI tool to manage Git hooks",
		Long: heredoc.Doc(`
				After installation go to your project directory
				and execute the following command:
				lefthook install
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(
		&options.Verbose, "verbose", "v", false, "verbose output",
	)
	rootCmd.PersistentFlags().BoolVar(
		&options.NoColors, "no-colors", false, "disable colored output",
	)

	rootCmd.Flags().BoolVarP(
		&options.Force, "force", "f", false,
		"DEPRECATED: reinstall hooks without checking config version",
	)
	rootCmd.Flags().BoolVarP(
		&options.Aggressive, "aggressive", "a", false,
		"DEPRECATED: remove all hooks from .git/hooks dir and install lefthook hooks",
	)

	for _, subcommand := range commands {
		rootCmd.AddCommand(subcommand(&options))
	}

	return rootCmd
}
