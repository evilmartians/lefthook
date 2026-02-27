package cmdtest

import (
	"io"
	"strings"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type Out struct {
	Command string
	Output  string
}

// OrderedCmd contains predefined list of commands and makes sure actual calls are the same.
type OrderedCmd struct {
	t    testing.TB
	outs []Out
	cnt  int
}

// WithoutEnvs simply does nothing.
func (c *OrderedCmd) WithoutEnvs(envs ...string) system.Command {
	return c
}

// Run makes sure command is executed correctly.
func (c *OrderedCmd) Run(command []string, root string, in io.Reader, out io.Writer, err io.Writer) error {
	c.t.Helper()
	defer func() { c.cnt += 1 }()

	cmd := strings.Join(command, " ")
	if len(c.outs) == 0 {
		c.t.Errorf("expected: no command, called: %s", cmd)
		return nil
	}

	checkCmd := c.outs[0]

	if checkCmd.Command != cmd {
		c.t.Errorf("%d) expected: '%s', called: '%s'", c.cnt, checkCmd.Command, cmd)
	}

	_, _ = out.Write([]byte(checkCmd.Output))
	c.outs = c.outs[1:]

	return nil
}
