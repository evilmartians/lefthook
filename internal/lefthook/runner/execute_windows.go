package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(opts ExecuteOptions) (*bytes.Buffer, error) {
	command := exec.Command(opts.args[0])
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(opts.args, " "),
	}

	rootDir, _ := filepath.Abs(opts.root)
	command.Dir = rootDir

	envList := make([]string, len(opts.env))
	for name, value := range opts.env {
		envList = append(envList, fmt.Sprintf("%s=%s", strings.ToUpper(name), value))
	}

	command.Env = append(os.Environ(), envList...)

	var out bytes.Buffer

	if opts.interactive {
		command.Stdout = os.Stdout
	} else {
		command.Stdout = &out
	}

	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return nil, err
	}

	defer command.Process.Kill()

	return &out, command.Wait()
}

func (e CommandExecutor) RawExecute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
