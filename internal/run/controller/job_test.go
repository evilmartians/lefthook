package controller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_withGitWorkTree(t *testing.T) {
	const repoRoot = "/repo"

	for name, tt := range map[string]struct {
		osGitDir, osGitWorkTree string
		env                     map[string]string
		want                    map[string]string
	}{
		"no-git-dir": {
			env:  map[string]string{"FOO": "bar"},
			want: map[string]string{"FOO": "bar"},
		},
		"git-dir-from-hook": {
			osGitDir: "/repo/.git/worktrees/wt",
			env:      map[string]string{},
			want:     map[string]string{"GIT_WORK_TREE": repoRoot},
		},
		"git-dir-from-hook-nil-env": {
			osGitDir: "/repo/.git/worktrees/wt",
			want:     map[string]string{"GIT_WORK_TREE": repoRoot},
		},
		"git-dir-from-job-env": {
			env: map[string]string{"GIT_DIR": "/other/.git"},
			want: map[string]string{
				"GIT_DIR":       "/other/.git",
				"GIT_WORK_TREE": repoRoot,
			},
		},
		"work-tree-from-hook": {
			osGitDir:      "/repo/.git/worktrees/wt",
			osGitWorkTree: "/elsewhere",
			env:           map[string]string{},
			want:          map[string]string{},
		},
		"work-tree-from-job-env": {
			osGitDir: "/repo/.git/worktrees/wt",
			env:      map[string]string{"GIT_WORK_TREE": "/elsewhere"},
			want:     map[string]string{"GIT_WORK_TREE": "/elsewhere"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("GIT_DIR", tt.osGitDir)
			t.Setenv("GIT_WORK_TREE", tt.osGitWorkTree)

			got := withGitWorkTree(tt.env, repoRoot)

			if !cmp.Equal(got, tt.want) {
				t.Errorf("withGitWorkTree() = %v, want %v", got, tt.want)
			}
		})
	}
}
