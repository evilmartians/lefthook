package cmdtest

import (
	"context"
	"io"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type SpyCmd struct {
	Commands []string
	callback func(cmd string, root string, out io.Writer) error
}

// NewSpyCmd returns executor that collects the called commands.
func NewSpyCmd(cb func(string, string, io.Writer) error) *SpyCmd {
	return &SpyCmd{
		Commands: make([]string, 0),
		callback: cb,
	}
}

// WithoutEnvs simply does nothing.
func (c *SpyCmd) WithoutEnvs(envs ...string) system.Command {
	return c
}

// Run makes sure command is executed correctly.
func (c *SpyCmd) Run(command []string, root string, in io.Reader, out io.Writer, err io.Writer) error {
	cmd := strings.Join(command, " ")
	c.Commands = append(c.Commands, cmd)

	if c.callback != nil {
		return c.callback(cmd, root, out)
	}

	return nil
}

func (c *SpyCmd) RunWithContext(_ context.Context, command []string, root string, in io.Reader, out io.Writer, err io.Writer) error {
	return c.Run(command, root, in, out, err)
}

func (c *SpyCmd) Reset() {
	c.Commands = []string{}
}
