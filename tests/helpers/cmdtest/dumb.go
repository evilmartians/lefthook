package cmdtest

import (
	"io"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type DumbCmd struct{}

// WithoutEnvs does nothing.
func (c *DumbCmd) WithoutEnvs(_ ...string) system.Command {
	return c
}

// Run does nothing.
func (c *DumbCmd) Run(_ []string, _ string, _ io.Reader, _ io.Writer, _ io.Writer) error {
	return nil
}
