package controller

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Controller) run(ctx context.Context, name string, follow bool, opts exec.Options) error {
	c.logger.Spinner.AddName(name)
	defer c.logger.Spinner.RemoveName(name)

	// Give every split command its own reader without consuming another job input.
	execute := func(out io.Writer) error {
		if !opts.UseStdin {
			return c.executor.Execute(ctx, opts, system.NullReader, out)
		}
		for _, command := range opts.Commands {
			split := opts
			split.Commands = []string{command}
			if err := c.executor.Execute(ctx, split, c.cachedStdin.Reader(), out); err != nil {
				return err
			}
		}
		return nil
	}

	if (follow || opts.Interactive) && c.logger.Enabled(logger.LogExecution) {
		c.logger.LogExecution(name, nil, nil)

		var out io.Writer
		if c.logger.Enabled(logger.LogExecutionOutput) {
			out = os.Stdout
		} else {
			out = io.Discard
		}

		return execute(out)
	}

	out := new(bytes.Buffer)
	err := execute(out)
	if out.Len() == 0 {
		c.logger.LogExecution(name, err, nil) // nothing to print: no empty line
	} else {
		c.logger.LogExecution(name, err, out)
	}

	return err
}
