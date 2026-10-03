package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_Pull(t *testing.T) {
	fs := afero.NewMemMapFs()
	errPull := errors.New("pull failed")

	for name, tt := range map[string]struct {
		err error
	}{
		"pulls": {},
		"pull-fails": {
			err: errPull,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{
				{Command: "git -C /repo pull --quiet", Err: tt.err},
			})
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.Pull("/repo")

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Pull() error = %v, want %v", err, tt.err)
			}
		})
	}
}
