// system package contains wrappers for OS interactions.
package system

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
)

type osCmd struct {
	excludeEnvs []string
}

var Cmd = osCmd{}

type Command interface {
	WithoutEnvs(...string) Command
	Run([]string, string, io.Reader, io.Writer, io.Writer) error
}

type CommandWithContext interface {
	RunWithContext(context.Context, []string, string, io.Reader, io.Writer, io.Writer) error
}

func (c osCmd) WithoutEnvs(envs ...string) Command {
	c.excludeEnvs = envs
	return c
}

func (c osCmd) Run(command []string, root string, in io.Reader, out io.Writer, errOut io.Writer) error {
	return c.RunWithContext(context.Background(), command, root, in, out, errOut)
}

// Run runs system command with LEFTHOOK=0 in order to prevent calling
// subsequent lefthook hooks.
func (c osCmd) RunWithContext(
	ctx context.Context,
	command []string,
	root string,
	in io.Reader,
	out io.Writer,
	errOut io.Writer,
) error {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if len(c.excludeEnvs) > 0 {
	loop:
		for _, env := range os.Environ() {
			for _, noenv := range c.excludeEnvs {
				if strings.HasPrefix(env, noenv) {
					continue loop
				}
			}
			cmd.Env = append(cmd.Env, env)
		}
		cmd.Env = append(cmd.Env, "LEFTHOOK=0")
	} else {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LEFTHOOK=0")
	}

	if len(root) > 0 {
		cmd.Dir = root
	}

	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errOut

	err := cmd.Run()

	b := log.DebugBuilder()
	b.Add("[lefthook] cmd:    ", strings.Join(command, " "))
	if len(root) > 0 {
		b.Add("[lefthook] dir:    ", root)
	}
	if err != nil {
		b.Add("[lefthook] error:  ", err)
	}
	b.Log()

	return err
}
