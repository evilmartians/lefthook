package wrapper_test

import (
	"errors"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_FilesByCommandRelative(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "a", []byte("content"), 0o644); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	errCommand := errors.New("command failed")

	command := "sh -c git ls-files"
	if runtime.GOOS == "windows" {
		command = "git ls-files"
	}

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want []string
		err  error
	}{
		"selects-existing-files": {
			out:  cmdtest.Out{Command: command, Output: "a\nmissing\n"},
			want: []string{"a"},
		},
		"command-fails": {
			out: cmdtest.Out{Command: command, Err: errCommand},
			err: errCommand,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.FilesByCommandRelative("git ls-files", "sub")

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.FilesByCommandRelative() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.FilesByCommandRelative() = %v, want %v", result, tt.want)
			}
		})
	}
}
