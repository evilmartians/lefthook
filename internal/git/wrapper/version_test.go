package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_Version(t *testing.T) {
	fs := afero.NewMemMapFs()
	errVersion := errors.New("version failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want string
		err  error
	}{
		"returns-trimmed-version": {
			out:  cmdtest.Out{Command: "git version", Output: "git version 2.50.0\n"},
			want: "git version 2.50.0",
		},
		"version-fails": {
			out: cmdtest.Out{Command: "git version", Output: "partial", Err: errVersion},
			err: errVersion,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.Version()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Version() error = %v, want %v", err, tt.err)
			}

			if result != tt.want {
				t.Errorf("wrapper.Version() = %q, want %q", result, tt.want)
			}
		})
	}
}
