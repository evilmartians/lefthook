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
	errList := errors.New("list failed")
	errDrop := errors.New("drop failed")
	list := "stash@{0}: lefthook auto backup\n" +
		"stash@{1}: On main: other\n" +
		"stash@{2}: lefthook auto backup\n"

	for name, tt := range map[string]struct {
		outs []cmdtest.Out
		err  error
	}{
		"drops-from-oldest": {
			outs: []cmdtest.Out{
				{Command: "git stash list", Output: list},
				{Command: "git stash drop --quiet -- stash@{2}"},
				{Command: "git stash drop --quiet -- stash@{0}"},
			},
		},
		"no-lefthook-stash": {
			outs: []cmdtest.Out{
				{Command: "git stash list", Output: "stash@{0}: On main: other\n"},
			},
		},
		"list-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash list", Err: errList},
			},
			err: errList,
		},
		"drop-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash list", Output: list},
				{Command: "git stash drop --quiet -- stash@{2}", Err: errDrop},
			},
			err: errDrop,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.DropStash()

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.DropStash() error = %v, want %v", err, tt.err)
			}
		})
	}
}
