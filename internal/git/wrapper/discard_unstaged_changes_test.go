package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_DiscardUnstagedChanges(t *testing.T) {
	fs := afero.NewMemMapFs()
	errCommand := errors.New("checkout failed")

	for name, tt := range map[string]struct {
		err error
	}{
		"succeeds": {},
		"checkout-fails": {
			err: errCommand,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{
				{Command: "git checkout --force -- a b", Err: tt.err},
			})
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.DiscardUnstagedChanges([]string{"a", "b"})

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.DiscardUnstagedChanges() error = %v, want %v", err, tt.err)
			}
		})
	}
}
