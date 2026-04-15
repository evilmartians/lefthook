//go:build !windows

package exec

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
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

func TestExecute_ConcurrentOutputIsolation_Interactive(t *testing.T) {
	// Mirrors how the controller runs parallel jobs: each goroutine gets
	// its own buffer. Output from concurrent commands must not leak across
	// buffers. Uses Interactive mode to take the direct exec path (no pty).
	const workers = 5
	const linesPerWorker = 20

	tmpDir := t.TempDir()

	bufs := make([]bytes.Buffer, workers)
	var wg sync.WaitGroup

	for i := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			label := fmt.Sprintf("WORKER-%d", id)
			cmd := fmt.Sprintf("for i in $(seq 1 %d); do echo '%s line '$i; sleep 0.01; done", linesPerWorker, label)
			opts := Options{
				Root:        tmpDir,
				Commands:    []string{cmd},
				Interactive: true,
			}
			err := CommandExecutor{}.Execute(context.Background(), opts, nil, &bufs[id])
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	for i := range workers {
		label := fmt.Sprintf("WORKER-%d", i)
		output := bufs[i].String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		assert.Len(t, lines, linesPerWorker, "%s should have %d lines", label, linesPerWorker)
		for _, line := range lines {
			assert.Contains(t, line, label, "buffer %d should only contain %s output, got: %s", i, label, line)
		}

		for j := range workers {
			if j == i {
				continue
			}
			other := fmt.Sprintf("WORKER-%d", j)
			assert.NotContains(t, output, other, "buffer %d should not contain %s output", i, other)
		}
	}
}
