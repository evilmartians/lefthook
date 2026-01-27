package controller

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Controller) run(ctx context.Context, name string, follow bool, opts exec.Options) error {
	log.SetName(name)
	defer log.UnsetName(name)

	// If the command does not explicitly `use_stdin` no input will be provided.
	var in io.Reader = system.NullReader
	if opts.UseStdin {
		in = c.cachedStdin
	}

	if (follow || opts.Interactive) && log.Settings.LogExecution() {
		log.Execution(name, nil, nil)

		var out io.Writer
		if log.Settings.LogExecutionOutput() {
			out = os.Stdout
		} else {
			out = io.Discard
		}

		return c.executor.Execute(ctx, opts, in, out)
	}

	out := new(bytes.Buffer)
	err := c.executor.Execute(ctx, opts, in, out)
	log.Execution(name, err, out)

	return err
}
