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
	"github.com/mattn/go-isatty"

	"github.com/evilmartians/lefthook/internal/log"
)

type CommandExecutor struct{}

func (e CommandExecutor) Execute(opts ExecuteOptions) (*bytes.Buffer, error) {
	stdin := os.Stdin
	if opts.interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			defer tty.Close()
			stdin = tty
		} else {
			log.Errorf("Couldn't enable TTY input: %s\n", err)
		}
	}

	command := exec.Command("sh", "-c", strings.Join(opts.args, " "))

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

	var out *bytes.Buffer

	if opts.interactive {
		command.Stdout = os.Stdout
		command.Stdin = stdin
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

		go func() { _, _ = io.Copy(p, stdin) }()

		out = bytes.NewBuffer(make([]byte, 0))
		_, _ = io.Copy(out, p)
	}

	defer func() { _ = command.Process.Kill() }()

	return out, command.Wait()
}

func (e CommandExecutor) RawExecute(command string, args ...string) (*bytes.Buffer, error) {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	return &out, cmd.Run()
}
