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

func TestWrapper_UnstagedDiffApplicable(t *testing.T) {
	patch := filepath.Join("/repo/.git/info", "lefthook-unstaged.patch")
	check := "git apply -v --whitespace=nowarn --recount --unidiff-zero --check -- " + patch

	for name, tt := range map[string]struct {
		patch *string
		outs  []cmdtest.Out
		want  bool
	}{
		"no-patch": {
			want: true,
		},
		"empty-patch": {
			patch: new(""),
			want:  true,
		},
		"check-succeeds": {
			patch: new("diff"),
			outs:  []cmdtest.Out{{Command: check}},
			want:  true,
		},
		"check-fails": {
			patch: new("diff"),
			outs:  []cmdtest.Out{{Command: check, Err: errors.New("apply failed")}},
			want:  false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.patch != nil {
				if err := afero.WriteFile(fs, patch, []byte(*tt.patch), 0o644); err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
			cmd := cmdtest.NewOrdered(t, append([]cmdtest.Out{pathsOut}, tt.outs...))
			w := wrapper.New(fs, cmd, loggertest.New())
			if _, err := w.Paths(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			result := w.UnstagedDiffApplicable()

			if result != tt.want {
				t.Errorf("wrapper.UnstagedDiffApplicable() = %v, want %v", result, tt.want)
			}
		})
	}
}
