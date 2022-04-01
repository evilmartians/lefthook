package lefthook

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunCommand(runner string, cmdRoot string) (*bytes.Buffer, bool, error) {
	runnerArgs := strings.Split(runner, " ")
	command := exec.Command(runnerArgs[0], runnerArgs[1:]...)
	if cmdRoot != "" {
		fullPath, _ := filepath.Abs(cmdRoot)
		command.Dir = fullPath
	}
	return RunPlainCommand(command)
}

func RunPlainCommand(command *exec.Cmd) (*bytes.Buffer, bool, error) {
	var commandOutput bytes.Buffer

	command.Stdout = &commandOutput
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	err := command.Start()
	if err != nil {
		return nil, false, err
	}
	return &commandOutput, true, command.Wait()
}
