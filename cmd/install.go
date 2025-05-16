package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
)

type install struct{}

func (install) New(opts *lefthook.Options) *cobra.Command {
	var a, force bool

	installCmd := cobra.Command{
		Use:               "install",
		Short:             "Write a basic configuration file in your project repository, or initialize the existing configuration",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _args []string) error {
			return lefthook.Install(opts, force)
		},
	}

	// To be dropped in next releases.
	installCmd.Flags().BoolVarP(
		&force, "force", "f", false,
		"overwrite .old hooks",
	)
	installCmd.Flags().BoolVarP(
		&a, "aggressive", "a", false,
		"use --force flag instead",
	)

	err := installCmd.Flags().MarkDeprecated("aggressive", "use --force flag instead")
	if err != nil {
		log.Warn("Unexpected error:", err)
	}

	return &installCmd
}
