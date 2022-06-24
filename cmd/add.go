package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

//go:embed add-doc.txt
var addDoc string

func newAddCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.AddArgs{}

	addCmd := cobra.Command{
		Use:     "add hook-name",
		Short:   "This command add a hook directory to a repository",
		Long:    addDoc,
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
