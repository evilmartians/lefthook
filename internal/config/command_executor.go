package config

import (
	"runtime"
)

// Executor is a general execution interface for implicit commands.
type Executor interface {
	Execute(args []string, root string) (string, error)
}

// commandExecutor implements execution of a skip checks passed in a `run` option.
type commandExecutor struct {
	exec Executor
}

// cmd runs plain string command in a subshell returning the success of it.
func (c *commandExecutor) cmd(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"powershell", "-Command", commandLine}
	} else {
		args = []string{"sh", "-c", commandLine}
	}

	_, err := c.exec.Execute(args, "")

	return err == nil
}
