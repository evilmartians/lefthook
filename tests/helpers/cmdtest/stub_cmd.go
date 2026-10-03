package cmdtest

import (
	"io"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type StubCmd struct{}

// NewStubCmd returns executor that does simply nothing.
func NewStubCmd() *StubCmd {
	return &StubCmd{}
}

// WithoutEnvs does nothing.
func (c *StubCmd) WithoutEnvs(_ ...string) system.Command {
	return c
}

// Run does nothing.
func (c *StubCmd) Run(_ []string, _ string, _ io.Reader, _ io.Writer, _ io.Writer) error {
	return nil
}
