package wrapper_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

// pathsOut answers wrapper.Paths() call with paths of a repository in /repo.
var pathsOut = cmdtest.Out{
	Command: "git rev-parse --path-format=absolute --show-toplevel --git-path hooks --git-path info --git-dir --git-common-dir",
	Output:  "/repo\n/repo/.git/hooks\n/repo/.git/info/\n/repo/.git\n/repo/.git\n",
}

func TestWrapper_Paths(t *testing.T) {
	fs := afero.NewMemMapFs()
	errPaths := errors.New("rev-parse failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want *wrapper.Paths
		err  error
	}{
		"returns-paths": {
			out: pathsOut,
			want: &wrapper.Paths{
				Root:      "/repo",
				Hooks:     "/repo/.git/hooks",
				Info:      filepath.Clean("/repo/.git/info/"),
				Git:       "/repo/.git",
				CommonGit: "/repo/.git",
			},
		},
		// In a linked worktree the --git-dir points to the worktree-specific
		// directory, while --git-common-dir points to the main repo's .git.
		// --git-path hooks resolves using the common dir, so Hooks == CommonGit/hooks.
		// ensureHooksPathUnset relies on this equality to avoid treating a normal
		// worktree as a command-scoped hooksPath override.
		"returns-paths-in-worktree": {
			out: cmdtest.Out{
				Command: pathsOut.Command,
				Output:  "/worktree\n/main/.git/hooks\n/main/.git/info/\n/main/.git/worktrees/wt1\n/main/.git\n",
			},
			want: &wrapper.Paths{
				Root:      "/worktree",
				Hooks:     "/main/.git/hooks",
				Info:      filepath.Clean("/main/.git/info/"),
				Git:       "/main/.git/worktrees/wt1",
				CommonGit: "/main/.git",
			},
		},
		"rev-parse-fails": {
			out: cmdtest.Out{Command: pathsOut.Command, Err: errPaths},
			err: errPaths,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.Paths()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Paths() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.Paths() = %v, want %v", result, tt.want)
			}
		})
	}
}
