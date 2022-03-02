//go:build !windows
// +build !windows

package lefthook

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"

	"github.com/evilmartians/lefthook/internal/log"
)

type WaitFunc func() error

func RunCommand(runner string, cmdRoot string) (*bytes.Buffer, bool, error) {
	command := exec.Command("sh", "-c", runner)
	if cmdRoot != "" {
		fullPath, _ := filepath.Abs(cmdRoot)
		command.Dir = fullPath
	}

	return RunPlainCommand(command)
}

func RunPlainCommand(command *exec.Cmd) (*bytes.Buffer, bool, error) {
	ptyOut, err := pty.Start(command)
	if err != nil {
		return nil, false, err
	}
	defer ptyOut.Close()

	commandOutput := new(bytes.Buffer)
	_, err = io.Copy(commandOutput, ptyOut)
	if err != nil {
		log.Debug(err)
	}

	return commandOutput, true, command.Wait()
}
