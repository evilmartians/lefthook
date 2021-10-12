package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var createDirs bool

var longAddCmd = heredoc.Doc(`
	This command will try to build the following structure in repository:
	├───.git
	│   └───hooks
	│       └───pre-commit // this executable will be added. Existed file with
	│                      // same name will be renamed to pre-commit.old
	(lefthook add this dirs if you run command with -d option)
	│
	├───.lefthook          // directory for project level hooks
	│   └───pre-commit     // directory with hooks executables
	├───.lefthook-local    // directory for personal hooks add it in .gitignore
	│   └───pre-commit
`)

func NewAddCmd(rootCmd *cobra.Command) {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "This command add a hook directory to a repository",
		Long:  longAddCmd,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addExecutor(args, appFs)
		},
	}

	addCmd.Flags().BoolVarP(&createDirs, "dirs", "d", false, "create directory for scripts")

	rootCmd.AddCommand(addCmd)
}

func addExecutor(args []string, appFs afero.Fs) error {
	// addHook
	// if createDirs
	//   addProjectHookDir
	//   addLocalHookDir

	return nil
}
