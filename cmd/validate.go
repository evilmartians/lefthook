package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

type validate struct{}

func (validate) New(opts *lefthook.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "validate",
		Short:   "Validate lefthook config",
		Long:    addDoc,
		Example: "lefthook validate",
		Args:    cobra.NoArgs,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return lefthook.Validate(opts)
		},
	}
}
