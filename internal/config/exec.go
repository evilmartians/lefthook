package config

import (
	"os/exec"
	"strings"
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
	parts := strings.Fields(commandLine)

	if len(parts) == 0 {
		return false
	}

	cmdName := parts[0]
	cmdArgs := parts[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	err := cmd.Run()

	return err == nil
}
