//go:build !windows
// +build !windows

package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
	"github.com/mattn/go-isatty"

	"github.com/evilmartians/lefthook/internal/log"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(opts Options, out io.Writer) error {
	in := os.Stdin
	if opts.Interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer tty.Close()
			in = tty
		} else {
			log.Errorf("Couldn't enable TTY input: %s\n", err)
		}
	}

	root, _ := filepath.Abs(opts.Root)
	envs := make([]string, len(opts.Env))
	for name, value := range opts.Env {
		envs = append(
			envs,
			fmt.Sprintf("%s=%s", strings.ToUpper(name), os.ExpandEnv(value)),
		)
	}

	// We can have one command split into separate to fit into shell command max length.
	// In this case we execute those commands one by one.
	for _, command := range opts.Commands {
		if err := e.executeOne(command, root, envs, opts.Interactive, in, out); err != nil {
			return err
		}
	}

	return nil
}

func (e CommandExecutor) RawExecute(command []string, out io.Writer) error {
	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e CommandExecutor) executeOne(cmdstr string, root string, envs []string, interactive bool, in io.Reader, out io.Writer) error {
	command := exec.Command("sh", "-c", cmdstr)
	command.Dir = root
	command.Env = append(os.Environ(), envs...)

	if interactive {
		command.Stdout = out
		command.Stdin = in
		command.Stderr = os.Stderr
		err := command.Start()
		if err != nil {
			return err
		}
	} else {
		p, err := pty.Start(command)
		if err != nil {
			return err
		}

		defer func() { _ = p.Close() }()

		go func() { _, _ = io.Copy(p, in) }()

		_, _ = io.Copy(out, p)
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}
