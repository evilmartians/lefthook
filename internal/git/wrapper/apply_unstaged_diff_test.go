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

func TestWrapper_ApplyUnstagedDiff(t *testing.T) {
	errApply := errors.New("apply failed")
	patch := filepath.Join("/repo/.git/info", "lefthook-unstaged.patch")
	patchAll := filepath.Join("/repo/.git/info", "lefthook-unstaged-all.patch")
	apply := "git apply -v --whitespace=nowarn --recount --unidiff-zero -- "

	for name, tt := range map[string]struct {
		all     bool
		patches map[string]string
		outs    []cmdtest.Out
		left    []string
		err     error
	}{
		"no-patch": {
			err: wrapper.ErrNoUnstagedDiff,
		},
		"empty-patch": {
			patches: map[string]string{patch: "", patchAll: "diff"},
		},
		"applies-patch": {
			patches: map[string]string{patch: "diff", patchAll: "diff"},
			outs:    []cmdtest.Out{{Command: apply + patch}},
		},
		"applies-all-patch": {
			all:     true,
			patches: map[string]string{patch: "diff", patchAll: "diff"},
			outs:    []cmdtest.Out{{Command: apply + patchAll}},
		},
		"apply-fails": {
			patches: map[string]string{patch: "diff", patchAll: "diff"},
			outs:    []cmdtest.Out{{Command: apply + patch, Err: errApply}},
			left:    []string{patch, patchAll},
			err:     errApply,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tt.patches {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
			cmd := cmdtest.NewOrdered(t, append([]cmdtest.Out{pathsOut}, tt.outs...))
			w := wrapper.New(fs, cmd, loggertest.New())
			if _, err := w.Paths(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			err := w.ApplyUnstagedDiff(tt.all)

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.ApplyUnstagedDiff() error = %v, want %v", err, tt.err)
			}

			var left []string
			for _, path := range []string{patch, patchAll} {
				if ok, _ := afero.Exists(fs, path); ok {
					left = append(left, path)
				}
			}
			if !cmp.Equal(left, tt.left) {
				t.Errorf("patches left = %v, want %v", left, tt.left)
			}
		})
	}
}
