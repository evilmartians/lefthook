package wrapper_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_SaveUnstagedDiff(t *testing.T) {
	fs := afero.NewMemMapFs()
	errDiff := errors.New("diff failed")
	diffArgs := "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output="
	diffFiles := diffArgs + filepath.Join("/repo/.git/info", "lefthook-unstaged.patch") + " -- a b"
	diffAll := diffArgs + filepath.Join("/repo/.git/info", "lefthook-unstaged-all.patch") + " --"

	// Paths of a git worktree "feature" of a repository in /repo.
	worktreePathsOut := cmdtest.Out{
		Command: pathsOut.Command,
		Output:  "/repo-feature\n/repo/.git/hooks\n/repo/.git/info/\n/repo/.git/worktrees/feature\n/repo/.git\n",
	}
	worktreeDiffFiles := diffArgs + filepath.Join("/repo/.git/info", "lefthook-unstaged.feature.patch") + " -- a b"
	worktreeDiffAll := diffArgs + filepath.Join("/repo/.git/info", "lefthook-unstaged-all.feature.patch") + " --"

	for name, tt := range map[string]struct {
		outs []cmdtest.Out
		err  error
	}{
		"saves-both-diffs": {
			outs: []cmdtest.Out{
				pathsOut,
				{Command: diffFiles},
				{Command: diffAll},
			},
		},
		"files-diff-fails": {
			outs: []cmdtest.Out{
				pathsOut,
				{Command: diffFiles, Err: errDiff},
			},
			err: errDiff,
		},
		"all-diff-fails": {
			outs: []cmdtest.Out{
				pathsOut,
				{Command: diffFiles},
				{Command: diffAll, Err: errDiff},
			},
			err: errDiff,
		},
		"worktree-saves-both-diffs": {
			outs: []cmdtest.Out{
				worktreePathsOut,
				{Command: worktreeDiffFiles},
				{Command: worktreeDiffAll},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())
			if _, err := w.Paths(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			err := w.SaveUnstagedDiff([]string{"a", "b"})

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.SaveUnstagedDiff() error = %v, want %v", err, tt.err)
			}
		})
	}
}
