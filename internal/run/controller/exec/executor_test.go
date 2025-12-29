package exec

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteWithTimeout(t *testing.T) {
	executor := CommandExecutor{}
	ctx := context.Background()

	t.Run("command completes before timeout", func(t *testing.T) {
		opts := Options{
			Commands: []string{"echo hello"},
			Timeout:  "5s",
		}
		out := new(bytes.Buffer)

		err := executor.Execute(ctx, opts, strings.NewReader(""), out)
		assert.NoError(t, err)
	})

	t.Run("command without timeout executes normally", func(t *testing.T) {
		opts := Options{
			Commands: []string{"echo hello"},
		}
		out := new(bytes.Buffer)

		err := executor.Execute(ctx, opts, strings.NewReader(""), out)
		assert.NoError(t, err)
	})

	t.Run("invalid timeout format returns error", func(t *testing.T) {
		opts := Options{
			Commands: []string{"echo hello"},
			Timeout:  "invalid",
		}
		out := new(bytes.Buffer)

		err := executor.Execute(ctx, opts, strings.NewReader(""), out)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timeout format")
	})
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "parse seconds",
			input:    "60s",
			expected: 60 * time.Second,
			wantErr:  false,
		},
		{
			name:     "parse minutes",
			input:    "5m",
			expected: 5 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "parse hours",
			input:    "2h",
			expected: 2 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "parse milliseconds",
			input:    "500ms",
			expected: 500 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "parse complex duration",
			input:    "1h30m45s",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
			wantErr:  false,
		},
		{
			name:     "parse zero",
			input:    "0s",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "invalid format - no unit",
			input:    "60",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - invalid unit",
			input:    "60x",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - just text",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Duration mismatch for input: %s", tt.input)
			}
		})
	}
}
