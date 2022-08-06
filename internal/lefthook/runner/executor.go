package runner

import (
	"bytes"
)

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(root string, args []string) (*bytes.Buffer, error)
}
