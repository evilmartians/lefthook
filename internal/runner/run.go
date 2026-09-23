package runner

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/runner/executor"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Runner) run(ctx context.Context, name string, follow bool, opts executor.Options) error {
	c.logger.Spinner.AddName(name)
	defer c.logger.Spinner.RemoveName(name)

	// If the command does not explicitly `use_stdin` no input will be provided.
	var in io.Reader = system.NullReader
	if opts.UseStdin {
		in = c.stdin()
	}

	if (follow || opts.Interactive) && c.logger.Enabled(logger.LogExecution) {
		c.logger.LogExecution(name, nil, nil)

		var out io.Writer
		if c.logger.Enabled(logger.LogExecutionOutput) {
			out = os.Stdout
		} else {
			out = io.Discard
		}

		return c.executor.Execute(ctx, opts, in, out)
	}

	out := new(bytes.Buffer)
	err := c.executor.Execute(ctx, opts, in, out)
	c.logger.LogExecution(name, err, out)

	return err
}
