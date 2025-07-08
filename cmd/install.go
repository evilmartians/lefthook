package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
)

type install struct{}

func (install) New(opts *command.Options) *cobra.Command {
	var force bool

	installCmd := cobra.Command{
		Use:               "install",
		Short:             "Write a basic configuration file in your project repository, or initialize the existing configuration",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _args []string) error {
			return command.Install(opts, force)
		},
	}

	// To be dropped in next releases.
	installCmd.Flags().BoolVarP(
		&force, "force", "f", false,
		"overwrite .old hooks",
	)

	return &installCmd
}
