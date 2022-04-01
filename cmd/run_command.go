//go:build !windows
// +build !windows

package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
)

type WaitFunc func() error

func RunCommand(runner string, cmdRoot string) (*bytes.Buffer, WaitFunc, error) {
	command := exec.Command("sh", "-c", runner)
	if cmdRoot != "" {
		fullPath, _ := filepath.Abs(cmdRoot)
		command.Dir = fullPath
	}

	return RunPlainCommand(command)
}

func RunPlainCommand(command *exec.Cmd) (*bytes.Buffer, WaitFunc, error) {
	ptyOut, err := pty.Start(command)

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptyOut, os.Stdin) }()
	commandOutput := bytes.NewBuffer(make([]byte, 0))
	_, _ = io.Copy(commandOutput, ptyOut)
	waitFunc := func() error {
		wErr := command.Wait()
		_ = ptyOut.Close()
		return wErr
	}
	return commandOutput, waitFunc, err
}
