package config

import (
	"runtime"

	"github.com/evilmartians/lefthook/internal/system"
)

type CommandExecutor interface {
	Cmd(commandLine string) bool
}

type executor struct {
	exec system.Executor
}

// NewExecutor returns an object that executes given commands in the OS.
func NewExecutor() CommandExecutor {
	return &executor{system.Executor{}}
}

// Cmd runs plain string command. It checks only exit code and returns bool value.
func (e *executor) Cmd(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"powershell", "-Command", commandLine}
	} else {
		args = []string{"sh", "-c", commandLine}
	}

	_, err := e.exec.Execute(args, "")

	return err == nil
}
