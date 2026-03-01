package controller

import (
	"context"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/command/replacer"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Controller) setup(
	ctx context.Context,
	opts Options,
	setupInstructions []*config.SetupInstruction,
) error {
	if len(setupInstructions) == 0 {
		return nil
	}

	log.StopSpinner()
	defer log.StartSpinner()

	replacer := replacer.New(c.git, "", "").
		AddTemplates(opts.Templates).
		AddGitArgs(opts.GitArgs)

	commands := make([]string, 0, len(setupInstructions))
	for _, instr := range setupInstructions {
		if err := replacer.Discover(instr.Run, nil); err != nil {
			return err
		}

		rawCommands, _ := replacer.ReplaceAndSplit(instr.Run, system.MaxCmdLen())
		commands = append(commands, rawCommands...)
	}

	log.LogSetup()

	return c.executor.Execute(ctx, exec.Options{Commands: commands}, system.NullReader, os.Stdout)
}
