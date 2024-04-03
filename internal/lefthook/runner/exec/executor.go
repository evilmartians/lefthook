package exec

import (
	"context"
	"io"
)

// Options contains the data that controls the execution.
type Options struct {
	Name, Root, FailText  string
	Commands              []string
	Env                   map[string]string
	Interactive, UseStdin bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(ctx context.Context, opts Options, out io.Writer) error
	RawExecute(ctx context.Context, command []string, out io.Writer) error
}
