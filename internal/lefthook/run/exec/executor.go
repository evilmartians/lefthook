package exec

import (
	"io"
)

// Options contains the data that controls the execution.
type Options struct {
	Name, Root, FailText string
	Commands             []string
	Env                  map[string]string
	Interactive          bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(opts Options, out io.Writer) error
	RawExecute(command []string, out io.Writer) error
}
