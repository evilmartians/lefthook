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
	command := exec.Command(opts.args[0])
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(opts.args, " "),
	}

	rootDir, _ := filepath.Abs(opts.root)
	command.Dir = rootDir

	envList := make([]string, len(opts.env))
	for name, value := range opts.env {
		envList = append(
			envList,
			fmt.Sprintf("%s=%s", strings.ToUpper(name), os.ExpandEnv(value)),
		)
	}

	command.Env = append(os.Environ(), envList...)

	if opts.interactive {
		command.Stdout = os.Stdout
	} else {
		command.Stdout = out
	}

	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return err
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}

func (e CommandExecutor) RawExecute(command []string, out io.Writer) error {
	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
