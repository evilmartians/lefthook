package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	commands = [...]func(*Options) *cobra.Command{
		NewVersionCmd,
		NewAddCmd,
		NewInstallCmd,
		NewUninstallCmd,
		NewRunCmd,
	}
)

func NewRootCmd() *cobra.Command {
	appOptions := &Options{
		fs: afero.NewOsFs(),
	}

	rootCmd := &cobra.Command{
		Use:   "lefthook",
		Short: "CLI tool to manage Git hooks",
		Long: heredoc.Doc(`
				After installation go to your project directory
				and execute the following command:
				lefthook install
		`),
	}

	rootCmd.PersistentFlags().BoolVarP(
		&appOptions.Verbose, "verbose", "v", false, "verbose output",
	)
	rootCmd.PersistentFlags().BoolVar(
		&appOptions.NoColors, "no-colors", false, "disable colored output",
	)

	// TODO: Drop deprecated options
	rootCmd.Flags().BoolVarP(
		&appOptions.Force, "force", "f", false,
		"DEPRECATED: reinstall hooks without checking config version",
	)
	rootCmd.Flags().BoolVarP(
		&appOptions.Aggressive, "aggressive", "a", false,
		"DEPRECATED: remove all hooks from .git/hooks dir and install lefthook hooks",
	)

	for _, subcommand := range commands {
		rootCmd.AddCommand(subcommand(appOptions))
	}

	return rootCmd
}
