//go:build !windows
// +build !windows

package exec

import (
	"context"
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

type executeArgs struct {
	in                    io.Reader
	out                   io.Writer
	envs                  []string
	root                  string
	interactive, useStdin bool
}

func (e CommandExecutor) Execute(ctx context.Context, opts Options, out io.Writer) error {
	var in io.Reader = nullReader{}
	if opts.UseStdin {
		in = os.Stdin
	}
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
	envs := make([]string, 0, len(opts.Env))
	for name, value := range opts.Env {
		envs = append(
			envs,
			fmt.Sprintf("%s=%s", strings.ToUpper(name), os.ExpandEnv(value)),
		)
	}

	args := &executeArgs{
		in:          in,
		out:         out,
		envs:        envs,
		root:        root,
		interactive: opts.Interactive,
		useStdin:    opts.UseStdin,
	}

	// We can have one command split into separate to fit into shell command max length.
	// In this case we execute those commands one by one.
	for _, command := range opts.Commands {
		if err := e.execute(ctx, command, args); err != nil {
			return err
		}
	}

	return nil
}

func (e CommandExecutor) RawExecute(ctx context.Context, command []string, out io.Writer) error {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e CommandExecutor) execute(ctx context.Context, cmdstr string, args *executeArgs) error {
	command := exec.CommandContext(ctx, "sh", "-c", cmdstr)
	command.Dir = args.root
	command.Env = append(os.Environ(), args.envs...)

	if args.interactive || args.useStdin {
		command.Stdout = args.out
		command.Stdin = args.in
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

		_, _ = io.Copy(args.out, p)
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}
