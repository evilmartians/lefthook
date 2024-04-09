package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-tty"
)

type CommandExecutor struct{}
type executeArgs struct {
	in   io.Reader
	out  io.Writer
	envs []string
	root string
}

func (e CommandExecutor) Execute(ctx context.Context, opts Options, out io.Writer) error {
	var in io.Reader = nullReader{}
	if opts.UseStdin {
		in = os.Stdin
	}
	if opts.Interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := tty.Open()
		if err == nil {
			defer tty.Close()
			in = tty.Input()
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

	args := &executeArgs{
		in:   in,
		out:  out,
		envs: envs,
		root: root,
	}

	for _, command := range opts.Commands {
		if err := e.execute(command, args); err != nil {
			return err
		}
	}

	return nil
}

func (e CommandExecutor) RawExecute(ctx context.Context, command []string, out io.Writer) error {
	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e CommandExecutor) execute(cmdstr string, args *executeArgs) error {
	cmdargs := strings.Split(cmdstr, " ")
	command := exec.Command(cmdargs[0])
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(cmdargs, " "),
	}
	command.Dir = args.root
	command.Env = append(os.Environ(), args.envs...)

	command.Stdout = args.out
	command.Stdin = args.in
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return err
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}
