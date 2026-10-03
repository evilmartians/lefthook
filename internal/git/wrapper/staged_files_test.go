package wrapper_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_StagedFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "a", []byte("content"), 0o644); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	errDiff := errors.New("diff failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want []string
		err  error
	}{
		"selects-existing-files": {
			out: cmdtest.Out{
				Command: "git diff --name-only --cached --diff-filter=ACMR",
				Output:  "a\nmissing\n",
			},
			want: []string{"a"},
		},
		"diff-fails": {
			out: cmdtest.Out{Command: "git diff --name-only --cached --diff-filter=ACMR", Err: errDiff},
			err: errDiff,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.StagedFiles()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.StagedFiles() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.StagedFiles() = %v, want %v", result, tt.want)
			}
		})
	}
}
