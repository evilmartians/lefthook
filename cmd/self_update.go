package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/updater"
)

type selfUpdate struct{}

func (selfUpdate) New(opts *lefthook.Options) *cobra.Command {
	var yes bool
	upgradeCmd := cobra.Command{
		Use:               "self-update",
		Short:             "Update lefthook executable",
		Example:           "lefthook self-update",
		ValidArgsFunction: cobra.NoFileCompletions,
		Args:              cobra.NoArgs,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return update(opts, yes)
		},
	}

	upgradeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "no prompt")
	upgradeCmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "force upgrade")
	upgradeCmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "show verbose logs")

	return &upgradeCmd
}

func update(opts *lefthook.Options, yes bool) error {
	if os.Getenv(lefthook.EnvVerbose) == "1" || os.Getenv(lefthook.EnvVerbose) == "true" {
		opts.Verbose = true
	}
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Verbose mode enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	go func() {
		<-signalChan
		cancel()
	}()

	return updater.New().SelfUpdate(ctx, yes, opts.Force)
}
