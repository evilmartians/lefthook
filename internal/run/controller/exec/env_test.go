package exec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithLefthookEnv(t *testing.T) {
	exe, err := os.Executable()
	require.NoError(t, err)
	exe, err = filepath.EvalSymlinks(exe)
	require.NoError(t, err)
	binDir := filepath.Dir(exe)

	t.Run("sets LEFTHOOK_BIN and prepends PATH", func(t *testing.T) {
		env := withLefthookEnv([]string{"FOO=bar", "PATH=/usr/bin"})
		assert.Contains(t, env, "FOO=bar")
		assert.Contains(t, env, lefthookBinEnv+"="+exe)

		var pathEntry string
		for _, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				pathEntry = e
				break
			}
		}
		require.NotEmpty(t, pathEntry)
		assert.True(t, strings.HasPrefix(pathEntry, "PATH="+binDir+string(os.PathListSeparator)))
		assert.Contains(t, pathEntry, "/usr/bin")
	})

	t.Run("does not override existing LEFTHOOK_BIN", func(t *testing.T) {
		env := withLefthookEnv([]string{lefthookBinEnv + "=/custom/lefthook", "PATH=/usr/bin"})
		assert.Contains(t, env, lefthookBinEnv+"=/custom/lefthook")
		assert.NotContains(t, env, lefthookBinEnv+"="+exe)
	})

	t.Run("does not duplicate bin dir in PATH", func(t *testing.T) {
		env := withLefthookEnv([]string{"PATH=" + binDir + ":/usr/bin"})
		pathCount := 0
		for _, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				pathCount++
			}
		}
		assert.Equal(t, 1, pathCount)
	})
}
