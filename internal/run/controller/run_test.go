package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		},
		{
			name:     "parse minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "parse hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "parse milliseconds",
			input:    "500ms",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "parse complex duration",
			input:    "1h30m45s",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:    "invalid format - no unit",
			input:   "60",
			wantErr: true,
		},
		{
			name:    "invalid format - invalid unit",
			input:   "60x",
			wantErr: true,
		},
		{
			name:    "invalid format - empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
