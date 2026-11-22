package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_DropStash(t *testing.T) {
	fs := afero.NewMemMapFs()
	errDrop := errors.New("drop failed")
	errStore := errors.New("store failed")
	storeOk := []cmdtest.Out{
		{Command: "git stash create", Output: "abc123"},
		{Command: "git update-ref refs/lefthook/backup/abc123 abc123 "},
	}

	for name, tt := range map[string]struct {
		// Commands of wrapper.StoreStash() call before wrapper.DropStash()
		store []cmdtest.Out
		outs  []cmdtest.Out
		err   error
	}{
		"drops-stored-only": {
			store: storeOk,
			outs: []cmdtest.Out{
				{Command: "git update-ref -d refs/lefthook/backup/abc123 abc123"},
			},
		},
		"drop-stored-fails": {
			store: storeOk,
			outs: []cmdtest.Out{
				{Command: "git update-ref -d refs/lefthook/backup/abc123 abc123", Err: errDrop},
			},
			err: errDrop,
		},
		"nothing-stored": {},
		"store-fails-drops-nothing": {
			store: []cmdtest.Out{
				{Command: "git stash create", Output: "abc123"},
				{Command: "git update-ref refs/lefthook/backup/abc123 abc123 ", Err: errStore},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, append(append([]cmdtest.Out{}, tt.store...), tt.outs...))
			w := wrapper.New(fs, cmd, loggertest.New())
			if len(tt.store) > 0 {
				_ = w.StoreStash()
			}

			err := w.DropStash()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.DropStash() error = %v, want %v", err, tt.err)
			}
		})
	}
}
