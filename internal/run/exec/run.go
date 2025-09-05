package exec

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
)

type RunOptions struct {
	Exec        Options
	CachedStdin io.Reader
	Follow      bool
}

func Run(ctx context.Context, executor Executor, opts *RunOptions) bool {
	name := opts.Exec.Name
	log.SetName(name)
	defer log.UnsetName(name)

	// If the command does not explicitly `use_stdin` no input will be provided.
	var in io.Reader = system.NullReader
	if opts.Exec.UseStdin {
		in = opts.CachedStdin
	}

	if (opts.Follow || opts.Exec.Interactive) && log.Settings.LogExecution() {
		log.Execution(name, nil, nil)

		var out io.Writer
		if log.Settings.LogExecutionOutput() {
			out = os.Stdout
		} else {
			out = io.Discard
		}

		err := executor.Execute(ctx, opts.Exec, in, out)

		return err == nil
	}

	out := new(bytes.Buffer)

	err := executor.Execute(ctx, opts.Exec, in, out)

	log.Execution(name, err, out)

	return err == nil
}
