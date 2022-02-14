package cmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/pkg/lefthook"
)

func NewAddCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.AddArgs{}

	addCmd := cobra.Command{
		Use:   "add hook-name",
		Short: "This command add a hook directory to a repository",
		Long: heredoc.Doc(`
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
    `),
		Example: "lefthook add pre-commit",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, hooks []string) error {
			args.Hook = hooks[0]
			return lefthook.Add(opts, &args)
		},
	}

	addCmd.Flags().BoolVarP(
		&args.CreateDirs, "dirs", "d", false, "create directory for scripts",
	)
	addCmd.Flags().BoolVarP(
		&args.Force, "force", "f", false, "overwrite .old hooks",
	)

	return &addCmd
}
