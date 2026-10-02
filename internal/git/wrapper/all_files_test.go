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

func TestWrapper_AllFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	for _, file := range []string{"a", "c d"} {
		if err := afero.WriteFile(fs, file, []byte("content"), 0o644); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	if err := fs.MkdirAll("dir", 0o755); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	errFiles := errors.New("ls-files failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want []string
		err  error
	}{
		"selects-existing-files": {
			out: cmdtest.Out{
				Command: "git ls-files --cached",
				Output:  "a\n\"c d\"\n  \ndir\nmissing\n",
			},
			want: []string{"a", "c d"},
		},
		"ls-files-fails": {
			out: cmdtest.Out{Command: "git ls-files --cached", Err: errFiles},
			err: errFiles,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.AllFiles()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.AllFiles() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.AllFiles() = %v, want %v", result, tt.want)
			}
		})
	}
}
