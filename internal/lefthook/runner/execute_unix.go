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
	"github.com/mattn/go-isatty"

	"github.com/evilmartians/lefthook/internal/log"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(root string, args []string, interactive bool) (*bytes.Buffer, error) {
	stdin := os.Stdin
	if interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer tty.Close()
			stdin = tty
		} else {
			log.Errorf("Couldn't enable TTY input: %s\n", err)
		}
	}
	command := exec.Command("sh", "-c", strings.Join(args, " "))
	rootDir, _ := filepath.Abs(root)
	command.Dir = rootDir

	p, err := pty.Start(command)
	if err != nil {
		return nil, err
	}

	defer func() { _ = p.Close() }()
	defer func() { _ = command.Process.Kill() }()

	go func() { _, _ = io.Copy(p, stdin) }()

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
