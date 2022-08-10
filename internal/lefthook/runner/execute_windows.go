package runner

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(root string, args []string) (*bytes.Buffer, error) {
	command := exec.Command(args[0])
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(args, " "),
	}

	rootDir, _ := filepath.Abs(root)
	command.Dir = rootDir

	var out bytes.Buffer

	command.Stdout = &out
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return nil, err
	}
	return &out, command.Wait()
}

func (e CommandExecutor) RawExecute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
