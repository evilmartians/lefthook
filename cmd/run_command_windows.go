package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WaitFunc func() error

func RunCommand(runner string, cmdRoot string) (*bytes.Buffer, WaitFunc, error) {
	runnerArgs := strings.Split(runner, " ")
	command := exec.Command(runnerArgs[0], runnerArgs[1:]...)
	if cmdRoot != "" {
		fullPath, _ := filepath.Abs(cmdRoot)
		command.Dir = fullPath
	}
	return RunPlainCommand(command)
}

func RunPlainCommand(command *exec.Cmd) (*bytes.Buffer, WaitFunc, error) {
	var commandOutput bytes.Buffer

	command.Stdout = &commandOutput
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	err := command.Start()
	return &commandOutput, command.Wait, err
}
