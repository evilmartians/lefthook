package controller

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

func Test_guard_wrap(t *testing.T) {
	for name, tt := range map[string]struct {
		stashUnstagedChanges bool
		failOnChanges        bool
		failOnChangesDiff    bool
		commands             []cmdtest.Out
		err                  error
	}{
		"just call": {
			stashUnstagedChanges: false,
			failOnChanges:        false,
			commands:             []cmdtest.Out{},
		},
		"failOnChanges=true no files": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
				{Command: "git status --short --porcelain -z", Output: ""},
			},
		},
		"failOnChanges=true no fail": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "0\n1\n"},
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "0\n1\n"},
			},
		},
		"failOnChanges=true fail with changeset different with diff": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			failOnChangesDiff:    true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "0\n1\n"},
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "2\n3\n"},
				{Command: "git diff --color -- file1 file2", Output: "diff --git a/file1 b/file1\n..."},
			},
			err: ErrFailOnChanges,
		},
		"failOnChanges=true fail with extra files without diff": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "0\n1\n"},
			},
			err: ErrFailOnChanges,
		},
		"stashUnstagedChanges=true no files": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
			},
		},
		"stashUnstagedChanges=true no unstaged": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "M  file1\x00M  file2\x00M  file3\x00"},
			},
		},
		"stashUnstagedChanges=true with partially staged": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "AM file1\x00 M file2\x00 A file3\x00"},
				{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
					filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
					" -- file1", Output: ""},
				{Command: "git stash create", Output: "<stash-hash>"},
				{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
				{Command: "git checkout --force -- file1", Output: ""},
				{Command: "git stash list", Output: "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
				{Command: "git stash drop --quiet 1", Output: ""},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo := gittest.NewRepositoryBuilder().
				Cmd(cmdtest.NewOrdered(t, tt.commands)).
				Fs(afero.NewMemMapFs()).
				Root("root").
				Build()
			repo.Setup()
			g := newGuard(repo, tt.stashUnstagedChanges, tt.failOnChanges, tt.failOnChangesDiff)

			var beenCalled bool
			err := g.wrap(func() { beenCalled = true })
			assert.Equal(tt.err, err)
			assert.Equal(true, beenCalled)
		})
	}
}
