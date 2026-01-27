package exec

import (
	"context"
	"io"
)

// Options contains the data that controls the execution.
type Options struct {
	Root                  string
	Commands              []string
	Env                   map[string]string
	Interactive, UseStdin bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(context.Context, Options, io.Reader, io.Writer) error
}
