package controller

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
)

func (c *Controller) setup(ctx context.Context, setupInstructions []*config.SetupInstruction) error {
	if len(setupInstructions) == 0 {
		return nil
	}

	commands := make([]string, 0, len(setupInstructions))
	for _, instr := range setupInstructions {
		commands = append(commands, instr.Run)
	}

	return c.run(ctx, "setup", true, exec.Options{
		Commands: commands,
	})
}
