package exec

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

func TestEnrichEnvWithLefthookBinary(t *testing.T) {
	const exe = "/tmp/lefthook-bin/lefthook"
	binDir := filepath.Dir(exe)

	for name, tt := range map[string]struct {
		env      []string
		contains []string
		excludes []string
		pathOnce bool
	}{
		"sets LEFTHOOK_BIN and prepends PATH": {
			env: []string{"FOO=bar", "PATH=/usr/bin"},
			contains: []string{
				"FOO=bar",
				lefthookBinEnv + "=" + exe,
			},
		},
		"does not override existing LEFTHOOK_BIN": {
			env: []string{lefthookBinEnv + "=/custom/lefthook", "PATH=/usr/bin"},
			contains: []string{
				lefthookBinEnv + "=/custom/lefthook",
			},
			excludes: []string{lefthookBinEnv + "=" + exe},
		},
		"does not duplicate bin dir in PATH": {
			env:      []string{"PATH=" + binDir + ":/usr/bin"},
			pathOnce: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			env := enrichEnvWithLefthookBinary(tt.env, exe)

			for _, want := range tt.contains {
				assert.Contains(t, env, want)
			}
			for _, omit := range tt.excludes {
				assert.NotContains(t, env, omit)
			}

			if name == "sets LEFTHOOK_BIN and prepends PATH" {
				pathEntry := envEntry(env, "PATH=")
				require.NotEmpty(t, pathEntry)
				assert.True(t, strings.HasPrefix(pathEntry, "PATH="+binDir+string(os.PathListSeparator)))
				assert.Contains(t, pathEntry, "/usr/bin")
			}

			if tt.pathOnce {
				assert.Equal(t, 1, strings.Count(strings.Join(env, "\n"), "PATH="))
			}
		})
	}
}

func TestWithLefthookEnv_resolutionErrors(t *testing.T) {
	for name, tt := range map[string]struct {
		executable func() (string, error)
		eval       func(string) (string, error)
		wantWarn   string
	}{
		"executable discovery error": {
			executable: func() (string, error) { return "", errors.New("no exe") },
			eval:       filepath.EvalSymlinks,
			wantWarn:   "Could not enrich hook PATH with lefthook binary: no exe",
		},
		"empty executable path": {
			executable: func() (string, error) { return "", nil },
			eval:       filepath.EvalSymlinks,
			wantWarn:   "Could not enrich hook PATH: empty executable path",
		},
		"symlink resolution error": {
			executable: func() (string, error) { return "/tmp/lefthook", nil },
			eval:       func(string) (string, error) { return "", errors.New("broken symlink") },
			wantWarn:   "Could not resolve lefthook binary path: broken symlink",
		},
	} {
		t.Run(name, func(t *testing.T) {
			oldExecutable := executablePath
			oldEval := evalSymlinks
			executablePath = tt.executable
			evalSymlinks = tt.eval
			t.Cleanup(func() {
				executablePath = oldExecutable
				evalSymlinks = oldEval
			})

			var buf bytes.Buffer
			log := logger.New(&buf).NewExecutionLogger()
			in := []string{"FOO=bar", "PATH=/usr/bin"}

			got := withLefthookEnv(in, log)

			assert.Equal(t, in, got)
			assert.Contains(t, buf.String(), tt.wantWarn)
		})
	}
}

func TestWithLefthookEnv_usesRunningBinary(t *testing.T) {
	exe, err := os.Executable()
	require.NoError(t, err)
	exe, err = filepath.EvalSymlinks(exe)
	require.NoError(t, err)

	env := withLefthookEnv([]string{"FOO=bar", "PATH=/usr/bin"}, nil)
	assert.Contains(t, env, lefthookBinEnv+"="+exe)
}

func envEntry(env []string, prefix string) string {
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return entry
		}
	}
	return ""
}
