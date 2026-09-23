//go:build !windows

package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/creack/pty"
	"github.com/mattn/go-isatty"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

type CommandExecutor struct {
	logger *logger.ExecutionLogger
}

type executeArgs struct {
	in                    io.Reader
	out                   io.Writer
	envs                  []string
	root                  string
	interactive, useStdin bool
}

func (e CommandExecutor) Execute(ctx context.Context, opts Options, in io.Reader, out io.Writer) error {
	if opts.Interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer func() {
				if cErr := tty.Close(); cErr != nil {
					e.logger.Warnf("Could not close TTY input: %s\n", cErr)
				}
			}()
			in = tty
		} else {
			e.logger.Errorf("Couldn't enable TTY input: %s\n", err)
		}
	}

	root, _ := filepath.Abs(opts.Root)
	envs := make([]string, 0, len(opts.Env))
	for name, value := range opts.Env {
		envs = append(
			envs,
			fmt.Sprintf("%s=%s", name, os.ExpandEnv(value)),
		)
	}
	if env, ok := colorEnv(e.logger, opts); ok {
		envs = append(envs, env)
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

func (e CommandExecutor) execute(ctx context.Context, cmdstr string, args *executeArgs) error {
	e.logger.Debug("[lefthook] run: ", cmdstr)
	command := exec.CommandContext(ctx, "sh", "-c", cmdstr)
	command.Dir = args.root
	command.Env = append(os.Environ(), args.envs...)

	switch {
	case args.interactive || args.useStdin:
		command.Stdout = args.out
		command.Stdin = args.in
		command.Stderr = os.Stderr
		err := command.Start()
		if err != nil {
			return err
		}
	case isatty.IsTerminal(os.Stdout.Fd()):
		return runInPTY(command, os.Stdout, args.out)
	default:
		// No pty available (sandbox, CI, pipe). Merge stderr into
		// stdout buffer to match pty behavior where both streams
		// go through the same device.
		//
		// Setpgid isolates the child from the parent's process group
		// so a SIGINT aimed at lefthook doesn't race with context
		// cancellation. Cancel kills the whole process group (negative
		// PID) so children like sleep(1) are cleaned up, matching the
		// session teardown that pty.Start (setsid) provides.
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = killProcessGroup(command)
		command.Stdout = args.out
		command.Stderr = args.out
		command.Stdin = args.in
		err := command.Start()
		if err != nil {
			return err
		}
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}

// runInPTY runs the command in a new PTY with the size of the given terminal
// and copies the command output to out.
func runInPTY(command *exec.Cmd, terminal *os.File, out io.Writer) error {
	// pty.Start makes the command a session leader (setsid), so it is also
	// the process group leader. Kill the whole group on cancel: children that
	// ignore SIGHUP survive the session teardown otherwise.
	command.Cancel = killProcessGroup(command)

	p, err := startWithInheritedSize(command, terminal)
	if err != nil {
		return err
	}

	defer func() { _ = p.Close() }()
	defer func() { _ = command.Process.Kill() }()

	_, _ = io.Copy(out, p)

	return command.Wait()
}

// killProcessGroup returns a cancel function that kills the process group of
// the command. The command must be the group leader (Setpgid or Setsid).
func killProcessGroup(command *exec.Cmd) func() error {
	return func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
}

func startWithInheritedSize(command *exec.Cmd, terminal *os.File) (*os.File, error) {
	size, err := pty.GetsizeFull(terminal)
	if err != nil {
		return nil, fmt.Errorf("get terminal size: %w", err)
	}

	p, err := pty.StartWithSize(command, size)
	if err != nil {
		return nil, fmt.Errorf("start command with PTY: %w", err)
	}

	return p, nil
}
