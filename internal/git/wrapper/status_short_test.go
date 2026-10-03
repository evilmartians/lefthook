package wrapper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_StatusShort(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := loggertest.New()
	cmd := cmdtest.NewOrdered(t, []cmdtest.Out{
		{
			Command: "git status --short --porcelain -z",
			Output: "RM new file\x00old-file\x00" +
				"M  staged\x00" +
				"MM staged but changed\x00" +
				"?? new.txt\x00" +
				"?? new-dir/\x00" +
				"RM new-file\x00old-file\x00" +
				"A  foo -> bar\x00" +
				"MM back\\slashes\x00" +
				"R  this is the new filename\x00R  this is really the old name, does it throw off the parser\x00" +
				"??  leading-space\x00",
		},
	})

	w := wrapper.New(fs, cmd, logger)

	result, err := w.StatusShort()
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}

	want := []wrapper.FileStatus{
		{Path: "new file", Index: 'R', Worktree: 'M'},
		{Path: "staged", Index: 'M', Worktree: ' '},
		{Path: "staged but changed", Index: 'M', Worktree: 'M'},
		{Path: "new.txt", Index: '?', Worktree: '?'},
		{Path: "new-dir/", Index: '?', Worktree: '?'},
		{Path: "new-file", Index: 'R', Worktree: 'M'},
		{Path: "foo -> bar", Index: 'A', Worktree: ' '},
		{Path: "back\\slashes", Index: 'M', Worktree: 'M'},
		{Path: "this is the new filename", Index: 'R', Worktree: ' '},
		{Path: " leading-space", Index: '?', Worktree: '?'},
	}
	if !cmp.Equal(result, want) {
		t.Errorf("wrapper.StatusShort = %v, want %v", result, want)
	}
}
