package cmd

import (
	_ "embed"
	"sort"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook"
)

//go:embed add-doc.txt
var addDoc string

type add struct{}

func (add) New(opts *lefthook.Options) *cobra.Command {
	args := lefthook.AddArgs{}

	addHookCompletions := func(cmd *cobra.Command, args []string, toComplete string) (ret []string, compDir cobra.ShellCompDirective) {
		compDir = cobra.ShellCompDirectiveNoFileComp
		if len(args) != 0 {
			return
		}
		ret = maps.Keys(config.AvailableHooks)
		sort.Strings(ret)
		return
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
