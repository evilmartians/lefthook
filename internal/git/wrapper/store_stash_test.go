package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_StoreStash(t *testing.T) {
	fs := afero.NewMemMapFs()
	errCreate := errors.New("create failed")
	errStore := errors.New("store failed")

	for name, tt := range map[string]struct {
		outs []cmdtest.Out
		err  error
	}{
		"stores-created-stash": {
			outs: []cmdtest.Out{
				{Command: "git stash create", Output: "abc123\n"},
				{Command: "git stash store --quiet --message lefthook auto backup abc123"},
			},
		},
		"create-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash create", Err: errCreate},
			},
			err: errCreate,
		},
		"store-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash create", Output: "abc123"},
				{Command: "git stash store --quiet --message lefthook auto backup abc123", Err: errStore},
			},
			err: errStore,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.StoreStash()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.StoreStash() error = %v, want %v", err, tt.err)
			}
		})
	}
}
