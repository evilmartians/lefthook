package config

import (
	"context"
	"io"
	"runtime"

	"github.com/evilmartians/lefthook/internal/system"
)

// commandExecutor implements execution of a skip checks passed in a `run` option.
type commandExecutor struct {
	cmd system.Command
}

// cmd runs plain string command in a subshell returning the success of it.
func (c *commandExecutor) execute(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"powershell", "-Command", commandLine}
	} else {
		args = []string{"sh", "-c", commandLine}
	}

	err := c.cmd.Run(context.Background(), args, "", system.NullReader, io.Discard)

	return err == nil
}
