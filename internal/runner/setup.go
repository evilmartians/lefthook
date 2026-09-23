package runner

import (
	"context"
	"io"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/runner/executor"
	"github.com/evilmartians/lefthook/v2/internal/runner/jobcmd/replacer"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Runner) setup(
	ctx context.Context,
	opts Options,
	setupInstructions []*config.SetupInstruction,
) error {
	if len(setupInstructions) == 0 {
		return nil
	}

	c.logger.Spinner.Stop()
	defer c.logger.Spinner.Start()

	replacer := replacer.New(c.git, c.logger, "", "").
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

	r, w := io.Pipe()
	c.logger.LogSetup(r)

	err := c.executor.Execute(ctx, executor.Options{Commands: commands}, system.NullReader, w)
	_ = w.Close()

	return err
}
