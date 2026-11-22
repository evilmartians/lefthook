package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_CommandHooksPath(t *testing.T) {
	fs := afero.NewMemMapFs()

	const showScopeCmd = "git config --show-scope --get core.hooksPath"

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want string
	}{
		"command-scoped returns value": {
			out:  cmdtest.Out{Command: showScopeCmd, Output: "command\t/tmp/hooks\n"},
			want: "/tmp/hooks",
		},
		"system-scoped returns empty": {
			out:  cmdtest.Out{Command: showScopeCmd, Output: "system\t/etc/git/hooks\n"},
			want: "",
		},
		"global-scoped returns empty": {
			out:  cmdtest.Out{Command: showScopeCmd, Output: "global\t/home/user/.config/git/hooks\n"},
			want: "",
		},
		"local-scoped returns empty": {
			out:  cmdtest.Out{Command: showScopeCmd, Output: "local\t.custom-hooks\n"},
			want: "",
		},
		"key not found (exit 1) returns empty": {
			out:  cmdtest.Out{Command: showScopeCmd, Err: errors.New("exit status 1")},
			want: "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			got := w.CommandHooksPath()
			if got != tt.want {
				t.Errorf("wrapper.CommandHooksPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
