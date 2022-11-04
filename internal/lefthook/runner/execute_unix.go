//go:build !windows
// +build !windows

package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(opts ExecuteOptions) (*bytes.Buffer, error) {
	command := exec.Command("sh", "-c", strings.Join(opts.args, " "))
	rootDir, _ := filepath.Abs(opts.root)
	command.Dir = rootDir

	envList := make([]string, len(opts.env))
	for name, value := range opts.env {
		envList = append(envList, fmt.Sprintf("%s=%s", strings.ToUpper(name), value))
	}

	command.Env = append(os.Environ(), envList...)

	var out *bytes.Buffer

	if opts.interactive {
		command.Stdout = os.Stdout
		command.Stdin = os.Stdin
		command.Stderr = os.Stderr
		err := command.Start()
		if err != nil {
			return nil, err
		}
	} else {
		p, err := pty.Start(command)
		if err != nil {
			return nil, err
		}

		defer func() { _ = p.Close() }()
		defer func() { _ = command.Process.Kill() }()

		go func() { _, _ = io.Copy(p, os.Stdin) }()

		out = bytes.NewBuffer(make([]byte, 0))
		_, _ = io.Copy(out, p)
	}

	defer command.Process.Kill()

	return out, command.Wait()
}

func (e CommandExecutor) RawExecute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
