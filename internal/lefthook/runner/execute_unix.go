//go:build !windows
// +build !windows

package runner

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"

	"github.com/evilmartians/lefthook/internal/log"
)

func Execute(name, root string, args []string) (*bytes.Buffer, error) {
	command := exec.Command("sh", "-c", strings.Join(args, " "))
	rootDir, _ := filepath.Abs(root)
	command.Dir = rootDir

	ptyOut, err := pty.Start(command)
	if err != nil {
		return nil, err
	}
	defer func() { _ = ptyOut.Close() }()

	out := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(out, ptyOut)
	if err != nil {
		log.Debug(err)
	}

	return out, command.Wait()
}
