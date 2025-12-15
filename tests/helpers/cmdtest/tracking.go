package cmdtest

import (
	"context"
	"io"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type TrackingCmd struct {
	Commands []string
	callback func(cmd string, root string, out io.Writer) error
}

// WithoutEnvs simply does nothing.
func (c *TrackingCmd) WithoutEnvs(envs ...string) system.Command {
	return c
}

// Run makes sure command is executed correctly.
func (c *TrackingCmd) Run(command []string, root string, in io.Reader, out io.Writer, err io.Writer) error {
	cmd := strings.Join(command, " ")
	c.Commands = append(c.Commands, cmd)

	if c.callback != nil {
		return c.callback(cmd, root, out)
	}

	return nil
}

func (c *TrackingCmd) RunWithContext(_ context.Context, command []string, root string, in io.Reader, out io.Writer, err io.Writer) error {
	return c.Run(command, root, in, out, err)
}

func (c *TrackingCmd) Reset() {
	c.Commands = []string{}
}
