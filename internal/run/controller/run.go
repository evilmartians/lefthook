package controller

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (c *Controller) run(ctx context.Context, name string, follow bool, opts exec.Options) (ok bool, failText string) {
	log.SetName(name)
	defer log.UnsetName(name)

	// Apply timeout if specified
	var timeoutDuration string
	if opts.Timeout != "" {
		timeout, err := parseDuration(opts.Timeout)
		if err != nil {
			log.Errorf("invalid timeout format '%s': %s\n", opts.Timeout, err)
			return false, ""
		}
		timeoutDuration = opts.Timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

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

		err := c.executor.Execute(ctx, opts, in, out)

		if err != nil && ctx.Err() == context.DeadlineExceeded {
			return false, "timeout (" + timeoutDuration + ")"
		}
		return err == nil, ""
	}

	out := new(bytes.Buffer)

	err := c.executor.Execute(ctx, opts, in, out)

	log.Execution(name, err, out)

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return false, "timeout (" + timeoutDuration + ")"
	}
	return err == nil, ""
}

// parseDuration parses a duration string (e.g., "60s", "5m", "1h30m").
func parseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}
