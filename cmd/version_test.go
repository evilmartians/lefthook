package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/command"
	"github.com/evilmartians/lefthook/internal/log"
)

func TestVersionCommand(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "version command without flags",
			args:     []string{},
			expected: "1.13.4",
		},
		{
			name:     "version command with full flag",
			args:     []string{"--full"},
			expected: "1.13.4 ", // Note: commit hash is empty in tests
		},
		{
			name:     "version command with full flag short form",
			args:     []string{"-f"},
			expected: "1.13.4 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Create version command
			opts := &command.Options{}
			versionCmd := version{}.New(opts)
			versionCmd.SetArgs(tt.args)

			// Execute the command
			err := versionCmd.Execute()
			assert.NoError(t, err)

			// Check output
			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestVersionCommand_FlagSetup(t *testing.T) {
	opts := &command.Options{}
	versionCmd := version{}.New(opts)

	// Test that the full flag exists and has correct properties
	fullFlag := versionCmd.Flags().Lookup("full")
	assert.NotNil(t, fullFlag)
	assert.Equal(t, "f", fullFlag.Shorthand)
	assert.Equal(t, "false", fullFlag.DefValue)
	assert.Equal(t, "full version with commit hash", fullFlag.Usage)

	// Test command properties
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Show lefthook version", versionCmd.Short)
}
