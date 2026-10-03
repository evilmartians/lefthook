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

func TestWrapper_State(t *testing.T) {
	gitPath := "/repo/.git"
	parents := `git show --no-patch --format="%P"`

	for name, tt := range map[string]struct {
		head  string
		files []string
		dirs  []string
		outs  []cmdtest.Out
		want  wrapper.State
	}{
		"no-state": {
			head: "ref: refs/heads/main\n",
			outs: []cmdtest.Out{{Command: parents, Output: "abc\n"}},
			want: wrapper.State{Branch: "main", State: ""},
		},
		"detached-head": {
			head: "abc\n",
			outs: []cmdtest.Out{{Command: parents, Output: "abc\n"}},
			want: wrapper.State{Branch: "", State: ""},
		},
		"no-head": {
			outs: []cmdtest.Out{{Command: parents, Output: "abc\n"}},
			want: wrapper.State{Branch: "", State: ""},
		},
		"merge": {
			head:  "ref: refs/heads/main\n",
			files: []string{"MERGE_HEAD"},
			want:  wrapper.State{Branch: "main", State: "merge"},
		},
		"rebase-merge": {
			head: "ref: refs/heads/feature/x\n",
			dirs: []string{"rebase-merge"},
			want: wrapper.State{Branch: "feature/x", State: "rebase"},
		},
		"rebase-apply": {
			head: "ref: refs/heads/main\n",
			dirs: []string{"rebase-apply"},
			want: wrapper.State{Branch: "main", State: "rebase"},
		},
		"merge-commit": {
			head: "ref: refs/heads/main\n",
			outs: []cmdtest.Out{{Command: parents, Output: "abc def\n"}},
			want: wrapper.State{Branch: "main", State: "merge-commit"},
		},
		"show-fails": {
			head: "ref: refs/heads/main\n",
			outs: []cmdtest.Out{{Command: parents, Err: errors.New("show failed")}},
			want: wrapper.State{Branch: "main", State: ""},
		},
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if len(tt.head) > 0 {
				if err := afero.WriteFile(fs, filepath.Join(gitPath, "HEAD"), []byte(tt.head), 0o644); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
			for _, file := range tt.files {
				if err := afero.WriteFile(fs, filepath.Join(gitPath, file), []byte("abc"), 0o644); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
			for _, dir := range tt.dirs {
				if err := fs.MkdirAll(filepath.Join(gitPath, dir), 0o755); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
			cmd := cmdtest.NewFakeCmd(t, append([]cmdtest.Out{pathsOut}, tt.outs...))
			w := wrapper.New(fs, cmd, loggertest.New())
			if _, err := w.Paths(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			result := w.State()

			if result != tt.want {
				t.Errorf("wrapper.State() = %v, want %v", result, tt.want)
			}
		})
	}
}
