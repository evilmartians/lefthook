package config

import (
	"os/exec"
	"runtime"
)

type Exec interface {
	Cmd(commandLine string) bool
}

type osExec struct{}

// NewOsExec returns an object that executes given commands in the OS.
func NewOsExec() Exec {
	return &osExec{}
}

// Cmd runs plain string command. It checks only exit code and returns bool value.
func (o *osExec) Cmd(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", commandLine)
	} else {
		cmd = exec.Command("sh", "-c", commandLine)
	}

	err := cmd.Run()

	return err == nil
}
