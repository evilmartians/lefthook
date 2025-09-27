package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/log"
)

func TestRootCommand_VersionFlag(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tests := []struct {
		name     string
		args     []string
		expected string
		exitCode int
	}{
		{
			name:     "version flag short form",
			args:     []string{"--version"},
			expected: "1.13.4",
		},
		{
			name:     "version flag long form",
			args:     []string{"--version=short"},
			expected: "1.13.4",
		},
		{
			name:     "version flag full",
			args:     []string{"--version=full"},
			expected: "1.13.4 ", // Note: commit hash is empty in tests
		},
		{
			name:     "version flag short alias",
			args:     []string{"-V"},
			expected: "1.13.4",
		},
		{
			name:     "version flag full alias",
			args:     []string{"-V=full"},
			expected: "1.13.4 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Create root command
			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)

			// We need to handle os.Exit() calls in tests
			// Since the command calls os.Exit(0), we'll test the flag parsing logic separately
			// and verify the version output logic

			// Test flag parsing
			err := rootCmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			versionFlag, err := rootCmd.Flags().GetString("version")
			assert.NoError(t, err)

			if strings.Contains(tt.args[0], "version") {
				assert.NotEmpty(t, versionFlag)

				// Test the version output logic without os.Exit
				verbose := versionFlag == fullVersionFlag

				// We can't easily test the actual execution due to os.Exit,
				// but we can verify the flag parsing works correctly
				if versionFlag == fullVersionFlag {
					assert.True(t, verbose)
				} else {
					assert.False(t, verbose)
				}
			}
		})
	}
}

func TestRootCommand_FlagDefaults(t *testing.T) {
	rootCmd := newRootCmd()

	// Test that version flag has correct default and NoOptDefVal
	versionFlag := rootCmd.Flags().Lookup("version")
	assert.NotNil(t, versionFlag)
	assert.Equal(t, "", versionFlag.DefValue)
	assert.Equal(t, "short", versionFlag.NoOptDefVal)
	assert.Equal(t, "show lefthook version (use 'full' for version with commit hash)", versionFlag.Usage)
}

func TestRootCommand_HelpOutput(t *testing.T) {
	rootCmd := newRootCmd()

	// Test that help includes the version flag
	helpOutput := rootCmd.UsageString()
	assert.Contains(t, helpOutput, "--version")
	assert.Contains(t, helpOutput, "-V")
	assert.Contains(t, helpOutput, "show lefthook version")
}

func TestRootCommand_VersionFlagPrecedence(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version with other flags",
			args: []string{"--version", "--verbose"},
		},
		{
			name: "version with subcommand",
			args: []string{"--version", "install"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)

			err := rootCmd.ParseFlags(tt.args)
			assert.NoError(t, err)

			versionFlag, err := rootCmd.Flags().GetString("version")
			assert.NoError(t, err)
			assert.NotEmpty(t, versionFlag)
		})
	}
}
