package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookExitError(t *testing.T) {
	err := &HookExitError{Code: 2}
	assert.Equal(t, 2, err.Code)
	assert.Equal(t, "", err.Error())
}
