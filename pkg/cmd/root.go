package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	appFs afero.Fs

	//TODO: move it to a configuration struct?

	Verbose  bool
	NoColors bool
)

func NewRootCmd() *cobra.Command {
	appFs = afero.NewOsFs()

	rootCmd := &cobra.Command{
		Use:   "lefthook",
		Short: "CLI tool to manage Git hooks",
		Long: heredoc.Doc(`
				After installation go to your project directory
				and execute the following command:
				lefthook install
		`),
	}

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&NoColors, "no-colors", false, "disable colored output")

	NewVersionCmd(rootCmd)
	NewAddCmd(rootCmd)
	NewInstallCmd(rootCmd)
	NewUninstallCmd(rootCmd)
	NewRunCmd(rootCmd)

	return rootCmd
}
