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
	// Err is returned instead of running the command, to simulate a failing git call.
	Err error
}

// OrderedCmd contains predefined list of commands and makes sure actual calls are the same.
type OrderedCmd struct {
	t    testing.TB
	outs []Out
	cnt  int
}

// NewOrdered returns executor that have the order defined in `outs`.
func NewOrdered(t testing.TB, outs []Out) *OrderedCmd {
	return &OrderedCmd{t: t, outs: outs}
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
		c.t.Errorf(`%d) "%v", want: "%v"`, c.cnt, cmd, checkCmd.Command)
	}

	_, _ = out.Write([]byte(checkCmd.Output))
	c.outs = c.outs[1:]

	return checkCmd.Err
}
