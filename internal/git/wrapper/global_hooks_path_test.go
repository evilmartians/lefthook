package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_GlobalHooksPath(t *testing.T) {
	fs := afero.NewMemMapFs()

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want string
	}{
		"returns-trimmed-path": {
			out:  cmdtest.Out{Command: "git config --global core.hooksPath", Output: ".hooks\n"},
			want: ".hooks",
		},
		"config-fails": {
			out: cmdtest.Out{Command: "git config --global core.hooksPath", Output: "partial", Err: errors.New("config failed")},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result := w.GlobalHooksPath()

			if result != tt.want {
				t.Errorf("wrapper.GlobalHooksPath() = %q, want %q", result, tt.want)
			}
		})
	}
}
