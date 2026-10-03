package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_UnsetLocalHooksPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	errCommand := errors.New("config failed")

	for name, tt := range map[string]struct {
		err error
	}{
		"succeeds": {},
		"config-fails": {
			err: errCommand,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{
				{Command: "git config --local --unset-all core.hooksPath", Err: tt.err},
			})
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.UnsetLocalHooksPath()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.UnsetLocalHooksPath() error = %v, want %v", err, tt.err)
			}
		})
	}
}
