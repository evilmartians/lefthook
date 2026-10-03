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

func TestWrapper_StagedFilesWithDeleted(t *testing.T) {
	fs := afero.NewMemMapFs()
	errDiff := errors.New("diff failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want []string
		err  error
	}{
		"keeps-deleted-files": {
			out: cmdtest.Out{
				Command: "git diff --name-only --cached --diff-filter=ACMRD",
				Output:  "a\n\"c d\"\n  \ndeleted\n",
			},
			want: []string{"a", "c d", "deleted"},
		},
		"diff-fails": {
			out: cmdtest.Out{Command: "git diff --name-only --cached --diff-filter=ACMRD", Err: errDiff},
			err: errDiff,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.StagedFilesWithDeleted()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.StagedFilesWithDeleted() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.StagedFilesWithDeleted() = %v, want %v", result, tt.want)
			}
		})
	}
}
