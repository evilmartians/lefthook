package exec

import (
	"context"
	"io"
)

// Options contains the data that controls the execution.
type Options struct {
	Name, Root            string
	Commands              []string
	Env                   map[string]string
	Interactive, UseStdin bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(ctx context.Context, opts Options, in io.Reader, out io.Writer) error
}
