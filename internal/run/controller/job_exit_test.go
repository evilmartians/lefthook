package controller

import (
	"errors"
	osexec "os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandExitCode(t *testing.T) {
	assert.Equal(t, 1, commandExitCode(errors.New("not an exit error")))

	err := osexec.Command("sh", "-c", "exit 42").Run()
	assert.Equal(t, 42, commandExitCode(err))
}
