package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
)

type validate struct{}

func (validate) New(opts *command.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "validate",
		Short:   "Validate lefthook config",
		Long:    addDoc,
		Example: "lefthook validate",
		Args:    cobra.NoArgs,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return command.Validate(opts)
		},
	}
}
