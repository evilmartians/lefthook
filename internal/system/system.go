// system package contains wrappers for OS interactions.
package system

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/evilmartians/lefthook/internal/log"
)

type osCmd struct{}

var Cmd = osCmd{}

type Command interface {
	Run([]string, string, io.Reader, io.Writer) error
}

type CommandWithContext interface {
	RunWithContext(context.Context, []string, string, io.Reader, io.Writer) error
}

// Run runs system command with LEFTHOOK=0 in order to prevent calling
// subsequent lefthook hooks.
func (c osCmd) RunWithContext(ctx context.Context, command []string, root string, in io.Reader, out io.Writer) error {
	log.Debug("[lefthook] cmd: ", command)

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	return run(cmd, root, in, out)
}

func (c osCmd) Run(command []string, root string, in io.Reader, out io.Writer) error {
	log.Debug("[lefthook] cmd: ", command)

	cmd := exec.Command(command[0], command[1:]...)

	return run(cmd, root, in, out)
}

func run(cmd *exec.Cmd, root string, in io.Reader, out io.Writer) error {
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")
	if len(root) > 0 {
		cmd.Dir = root
		log.Debug("[lefthook] dir: ", root)
	}

	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	log.Debug("[lefthook] err: ", err)

	return err
}
