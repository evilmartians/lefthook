package runner

import (
	"io"
)

// ExecutorOptions contains the options that control the execution.
type ExecuteOptions struct {
	name, root, failText string
	args                 []string
	env                  map[string]string
	interactive          bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(opts ExecuteOptions, out io.Writer) error
	RawExecute(command []string, out io.Writer) error
}
