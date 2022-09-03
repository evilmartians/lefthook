//go:build !windows
// +build !windows

package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(root string, args []string) (*bytes.Buffer, error) {
	command := exec.Command("sh", "-c", strings.Join(args, " "))
	rootDir, _ := filepath.Abs(root)
	command.Dir = rootDir

	p, err := pty.Start(command)
	if err != nil {
		return nil, err
	}

	defer func() { _ = p.Close() }()
	defer command.Process.Kill()

	go func() { io.Copy(p, os.Stdin) }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, _ = io.Copy(out, p)

	return out, command.Wait()
}

func (e CommandExecutor) RawExecute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
