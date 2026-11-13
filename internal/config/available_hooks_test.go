package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookUsesStagedFiles(t *testing.T) {
	for _, hook := range []string{"pre-commit"} {
		assert.True(t, HookUsesStagedFiles(hook), hook)
	}

	for _, hook := range []string{"pre-merge-commit", "pre-push", "commit-msg", "post-commit"} {
		assert.False(t, HookUsesStagedFiles(hook), hook)
	}
}

func TestHookFallsBackToStagedFiles(t *testing.T) {
	// pre-merge-commit falls back to the staged files like pre-commit, but
	// stage_fixed stays limited to pre-commit.
	for _, hook := range []string{"pre-commit", "pre-merge-commit"} {
		assert.True(t, HookFallsBackToStagedFiles(hook), hook)
	}

	for _, hook := range []string{"pre-push", "commit-msg", "post-commit", "pre-rebase"} {
		assert.False(t, HookFallsBackToStagedFiles(hook), hook)
	}
}
