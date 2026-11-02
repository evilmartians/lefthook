package exectest

import (
	"context"
	"io"

	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
)

type ExecutorFunc func(context.Context, exec.Options, io.Reader, io.Writer) error

func (f ExecutorFunc) Execute(ctx context.Context, opts exec.Options, in io.Reader, out io.Writer) error {
	return f(ctx, opts, in, out)
}
