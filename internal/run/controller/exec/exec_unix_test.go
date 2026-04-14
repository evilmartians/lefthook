//go:build !windows

package exec

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	tmpDir := t.TempDir()
	// Resolve symlinks so pwd -P output matches (macOS: /var -> /private/var).
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", tmpDir, err)
	}

	for name, tt := range map[string]struct {
		opts       Options
		wantOut    string
		wantNotOut string
		wantErr    bool
	}{
		"captures stdout": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"echo hello"},
			},
			wantOut: "hello",
		},
		"captures stderr": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"echo error-output >&2"},
			},
			wantOut: "error-output",
		},
		"propagates env": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"echo $TEST_EXEC_VAR"},
				Env:      map[string]string{"TEST_EXEC_VAR": "propagated"},
			},
			wantOut: "propagated",
		},
		"respects working directory": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"pwd -P"},
			},
			wantOut: realTmpDir,
		},
		"returns error on nonzero exit": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"exit 1"},
			},
			wantErr: true,
		},
		"runs multiple commands": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"echo first-cmd", "echo second-cmd"},
			},
			wantOut: "second-cmd",
		},
		"stops at first failing command": {
			opts: Options{
				Root:     tmpDir,
				Commands: []string{"exit 1", "echo should-not-run"},
			},
			wantErr:    true,
			wantNotOut: "should-not-run",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := CommandExecutor{}.Execute(context.Background(), tt.opts, nil, &buf)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantOut != "" {
				assert.Contains(t, buf.String(), tt.wantOut)
			}
			if tt.wantNotOut != "" {
				assert.NotContains(t, buf.String(), tt.wantNotOut)
			}
		})
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	opts := Options{
		Root:     tmpDir,
		Commands: []string{"sleep 10"},
	}

	start := time.Now()
	err := CommandExecutor{}.Execute(ctx, opts, nil, &buf)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, elapsed, 5*time.Second, "should return promptly after context cancellation")
}

func TestExecute_UseStdin(t *testing.T) {
	tmpDir := t.TempDir()
	var buf bytes.Buffer
	in := strings.NewReader("from-stdin\n")

	opts := Options{
		Root:     tmpDir,
		Commands: []string{"cat"},
		UseStdin: true,
	}

	err := CommandExecutor{}.Execute(context.Background(), opts, in, &buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "from-stdin")
}
