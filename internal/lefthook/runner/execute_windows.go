package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(opts ExecuteOptions, out io.Writer) error {
	root, _ := filepath.Abs(opts.root)
	envs := make([]string, len(opts.env))
	for name, value := range opts.env {
		envs = append(
			envs,
			fmt.Sprintf("%s=%s", strings.ToUpper(name), os.ExpandEnv(value)),
		)
	}

	for _, args := range opts.commands {
		if err := e.executeOne(args, root, envs, opts.interactive, os.Stdin, out); err != nil {
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

func (e CommandExecutor) executeOne(args []string, root string, envs []string, interactive bool, in io.Reader, out io.Writer) error {
	command := exec.Command(args[0])
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(args, " "),
	}
	command.Dir = root
	command.Env = append(os.Environ(), envs...)

	if interactive {
		command.Stdout = os.Stdout
	} else {
		command.Stdout = out
	}

	command.Stdin = in
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return err
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}
