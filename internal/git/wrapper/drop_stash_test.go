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
	errStore := errors.New("store failed")
	list := "stash@{0} hash lefthook auto backup\n" +
		"stash@{1} hash On main: other\n" +
		"stash@{2} hash lefthook auto backup\n"
	listStored := "stash@{0} abc123 lefthook auto backup\n" +
		"stash@{1} def456 On main: other\n" +
		"stash@{2} 789fed lefthook auto backup\n"
	storeOk := []cmdtest.Out{
		{Command: "git stash create", Output: "abc123"},
		{Command: "git stash store --quiet --message lefthook auto backup abc123"},
	}

	for name, tt := range map[string]struct {
		// Commands of wrapper.StoreStash() call before wrapper.DropStash()
		store []cmdtest.Out
		outs  []cmdtest.Out
		err   error
	}{
		"drops-from-oldest": {
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: list},
				{Command: "git stash drop --quiet -- stash@{2}"},
				{Command: "git stash drop --quiet -- stash@{0}"},
			},
		},
		"no-lefthook-stash": {
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: "stash@{0} hash On main: other\n"},
			},
		},
		"list-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Err: errList},
			},
			err: errList,
		},
		"drop-fails": {
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: list},
				{Command: "git stash drop --quiet -- stash@{2}", Err: errDrop},
			},
			err: errDrop,
		},
		"drops-stored-only": {
			store: storeOk,
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: listStored},
				{Command: "git stash drop --quiet -- stash@{0}"},
			},
		},
		"stored-not-listed": {
			store: storeOk,
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: list},
			},
		},
		"drop-stored-fails": {
			store: storeOk,
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: listStored},
				{Command: "git stash drop --quiet -- stash@{0}", Err: errDrop},
			},
			err: errDrop,
		},
		"store-fails-drops-all": {
			store: []cmdtest.Out{
				{Command: "git stash create", Output: "abc123"},
				{Command: "git stash store --quiet --message lefthook auto backup abc123", Err: errStore},
			},
			outs: []cmdtest.Out{
				{Command: "git stash list --format=%gd %H %gs", Output: listStored},
				{Command: "git stash drop --quiet -- stash@{2}"},
				{Command: "git stash drop --quiet -- stash@{0}"},
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
