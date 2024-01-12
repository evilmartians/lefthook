package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook"
)

//go:embed add-doc.txt
var addDoc string

func newAddCmd(opts *lefthook.Options) *cobra.Command {
	args := lefthook.AddArgs{}

	addHookCompletions := func(cmd *cobra.Command, args []string, toComplete string) (ret []string, compDir cobra.ShellCompDirective) {
		compDir = cobra.ShellCompDirectiveNoFileComp
		if len(args) != 0 {
			return
		}
		ret = config.AvailableHooks[:]
		return
	}

	addCmd := cobra.Command{
		Use:               "add hook-name",
		Short:             "This command adds a hook directory to a repository",
		Long:              addDoc,
		Example:           "lefthook add pre-commit",
		ValidArgsFunction: addHookCompletions,
		Args:              cobra.MinimumNArgs(1),
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
