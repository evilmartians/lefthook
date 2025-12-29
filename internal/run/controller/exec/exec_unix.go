//go:build !windows

package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creack/pty"
	"github.com/mattn/go-isatty"

	"github.com/evilmartians/lefthook/v2/internal/log"
)

type CommandExecutor struct{}

type executeArgs struct {
	in                    io.Reader
	out                   io.Writer
	envs                  []string
	root                  string
	interactive, useStdin bool
}

func (e CommandExecutor) Execute(ctx context.Context, opts Options, in io.Reader, out io.Writer) error {
	// Apply timeout if specified
	if opts.Timeout != "" {
		timeout, err := parseDuration(opts.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format '%s': %w", opts.Timeout, err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if opts.Interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer func() {
				if cErr := tty.Close(); cErr != nil {
					log.Warnf("Could not close TTY input: %s\n", cErr)
				}
			}()
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
			fmt.Sprintf("%s=%s", name, os.ExpandEnv(value)),
		)
	}
	switch log.Colors() {
	case log.ColorOn:
		envs = append(envs, "CLICOLOR_FORCE=true")
	case log.ColorOff:
		envs = append(envs, "NO_COLOR=true")
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
	log.Debug("[lefthook] run: ", cmdstr)
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

// parseDuration parses a duration string (e.g., "60s", "5m", "1h30m").
func parseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}
