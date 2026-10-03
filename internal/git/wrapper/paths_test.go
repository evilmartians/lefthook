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
	Command: "git rev-parse --path-format=absolute --show-toplevel --git-path hooks --git-path info --git-dir",
	Output:  "/repo\n/repo/.git/hooks\n/repo/.git/info/\n/repo/.git\n",
}

func TestWrapper_Paths(t *testing.T) {
	errPaths := errors.New("rev-parse failed")

	for name, tt := range map[string]struct {
		out   cmdtest.Out
		files []string
		want  *wrapper.Paths
		err   error
	}{
		"returns-paths": {
			out: pathsOut,
			want: &wrapper.Paths{
				Root:  "/repo",
				Hooks: "/repo/.git/hooks",
				Info:  filepath.Clean("/repo/.git/info/"),
				Git:   "/repo/.git",
			},
		},
		"rev-parse-fails": {
			out: cmdtest.Out{Command: pathsOut.Command, Err: errPaths},
			err: errPaths,
		},
		"falls-back-to-git-dir-parent-when-toplevel-is-elsewhere": {
			out: cmdtest.Out{
				Command: pathsOut.Command,
				Output: "/Users/user/.cache/uv/git-v0/db/001c65abee8f6b9f\n" +
					"/Users/user/Developer/actual-git-repo/.git/hooks\n" +
					"/Users/user/Developer/actual-git-repo/.git/info/\n" +
					"/Users/user/Developer/actual-git-repo/.git\n",
			},
			want: &wrapper.Paths{
				Root:  "/Users/user/Developer/actual-git-repo",
				Hooks: "/Users/user/Developer/actual-git-repo/.git/hooks",
				Info:  filepath.Clean("/Users/user/Developer/actual-git-repo/.git/info/"),
				Git:   "/Users/user/Developer/actual-git-repo/.git",
			},
		},
		"keeps-toplevel-for-worktree-with-gitlink": {
			out: cmdtest.Out{
				Command: pathsOut.Command,
				Output: "/repo-wt\n/main/.git/worktrees/wt/hooks\n" +
					"/main/.git/worktrees/wt/info/\n/main/.git/worktrees/wt\n",
			},
			files: []string{"/repo-wt/.git"},
			want: &wrapper.Paths{
				Root:  "/repo-wt",
				Hooks: "/main/.git/worktrees/wt/hooks",
				Info:  filepath.Clean("/main/.git/worktrees/wt/info/"),
				Git:   "/main/.git/worktrees/wt",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for _, file := range tt.files {
				if err := afero.WriteFile(fs, file, []byte("gitdir: /main/.git"), 0o644); err != nil {
					t.Fatalf("failed to prepare fs: %v", err)
				}
			}
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
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
