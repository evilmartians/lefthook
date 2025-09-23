package cmd

import (
	_ "embed"
	"maps"
	"slices"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/config"
)

//go:embed add-doc.txt
var addDoc string

type add struct{}

func (add) New(opts *command.Options) *cobra.Command {
	args := command.AddArgs{}

	addHookCompletions := func(cmd *cobra.Command, args []string, toComplete string) (ret []string, compDir cobra.ShellCompDirective) {
		compDir = cobra.ShellCompDirectiveNoFileComp
		if len(args) != 0 {
			return ret, compDir
		}
		ret = slices.Sorted(maps.Keys(config.AvailableHooks))
		return ret, compDir
	}

	addCmd := cobra.Command{
		Use:               "add hook-name",
		Short:             "This command adds a hook directory to a repository",
		Long:              addDoc,
		Example:           "lefthook add pre-commit",
		ValidArgsFunction: addHookCompletions,
		Args:              cobra.ExactArgs(1),
		RunE: func(_cmd *cobra.Command, hooks []string) error {
			args.Hook = hooks[0]
			return command.Add(opts, &args)
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
