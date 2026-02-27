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
			err: &FailOnChangesError{[]string{"file1", "file2"}},
		},
		"failOnChanges=true fail with extra files without diff": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
				{Command: "git status --short --porcelain -z", Output: " M file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "0\n1\n"},
			},
			err: &FailOnChangesError{[]string{"file1", "file2"}},
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
		"stashUnstagedChanges=true failOnChanges=true with partially staged no hook changes": {
			stashUnstagedChanges: true,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "AM file1\x00 M file2\x00"},
				{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
					filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
					" -- file1", Output: ""},
				{Command: "git stash create", Output: "<stash-hash>"},
				{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
				{Command: "git checkout --force -- file1", Output: ""},
				{Command: "git status --short --porcelain -z", Output: "A file1\x00"},
				// job run
				{Command: "git status --short --porcelain -z", Output: "A file1\x00"},
				{Command: "git stash list", Output: "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
				{Command: "git stash drop --quiet 1", Output: ""},
			},
		},
		// "stashUnstagedChanges=true failOnChanges=true with partially staged and hook changes": {
		// 	stashUnstagedChanges: true,
		// 	failOnChanges:        true,
		// 	commands: []cmdtest.Out{
		// 		{Command: "git status --short --porcelain -z", Output: "AM file1\x00"},
		// 		{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
		// 			filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
		// 			" -- file1", Output: ""},
		// 		{Command: "git stash create", Output: "<stash-hash>"},
		// 		{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
		// 		{Command: "git checkout --force -- file1", Output: ""},
		// 		{Command: "git status --short --porcelain -z", Output: "A file1\x00"},
		// 		{Command: "git hash-object -- file1", Output: "hash1\n"},
		// 		// run jobs
		// 		{Command: "git status --short --porcelain -z", Output: "A  file1\x00 M file2\x00"},
		// 		{Command: "git hash-object -- file1 file2", Output: "hash1\nhash2\n"},
		// 		{Command: "git status --short --porcelain -z", Output: "A  file1\x00 M file2\x00"},
		// 		{Command: "git hash-object -- file1 file2", Output: "hash1\nhash3\n"},
		// 		{Command: "git stash list", Output: "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
		// 		{Command: "git stash drop --quiet 1", Output: ""},
		// 	},
		// 	err: &ErrFailOnChanges{},
		// },
		"stashUnstagedChanges=true failOnChanges=true with partially staged and hook changes with diff": {
			stashUnstagedChanges: true,
			failOnChanges:        true,
			failOnChangesDiff:    true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "AM file1\x00"},
				{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
					filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
					" -- file1", Output: ""},
				{Command: "git stash create", Output: "<stash-hash>"},
				{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
				{Command: "git checkout --force -- file1", Output: ""},
				{Command: "git status --short --porcelain -z", Output: "A  file1\x00"},
				{Command: "git hash-object -- file1", Output: "hash1\n"},
				// job run
				{Command: "git status --short --porcelain -z", Output: "A  file1\x00 M file2\x00"},
				{Command: "git hash-object -- file1 file2", Output: "hash1\nhash3\n"},
				{Command: "git diff --color -- file2", Output: "diff --git a/file2 b/file2\n..."},
				{Command: "git checkout --force -- file2", Output: ""},
				{Command: "git stash list", Output: "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
				{Command: "git stash drop --quiet 1", Output: ""},
			},
			err: &FailOnChangesError{[]string{"file2"}},
		},
		"failOnChanges=true with deleted file no change": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				// Deleted file in before and after - same state, no change
				{Command: "git status --short --porcelain -z", Output: "D  file1\x00"},
				{Command: "git status --short --porcelain -z", Output: "D  file1\x00"},
			},
		},
		"failOnChanges=true with directory": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				// Directory in before() - marked as "directory"
				{Command: "git status --short --porcelain -z", Output: "?? dir/\x00"},
				// Directory still there in after() - same state, no change
				{Command: "git status --short --porcelain -z", Output: "?? dir/\x00"},
			},
		},
		"failOnChanges=true with changeset error in before": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				// Changeset() error in before() - empty output simulates error
				{Command: "git status --short --porcelain -z", Output: ""},
				{Command: "git status --short --porcelain -z", Output: ""},
			},
		},
		// "stashUnstagedChanges=true failOnChanges=true with changeset error after stashing": {
		// 	stashUnstagedChanges: true,
		// 	failOnChanges:        true,
		// 	commands: []cmdtest.Out{
		// 		// First Changeset() in before() - has file1
		// 		{Command: "git status --short --porcelain -z", Output: "AM file1\x00"},
		// 		{Command: "git hash-object -- file1", Output: "hash1\n"},
		// 		// PartiallyStagedFiles() - uses statusShortOnce
		// 		{Command: "git status --short --porcelain -z", Output: "AM file1\x00"},
		// 		{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
		// 			filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
		// 			" -- file1", Output: ""},
		// 		{Command: "git stash create", Output: "<stash-hash>"},
		// 		{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
		// 		{Command: "git checkout --force -- file1", Output: ""},
		// 		// Second Changeset() in before() after stashing - empty (simulates error/no files)
		// 		{Command: "git status --short --porcelain -z", Output: ""},
		// 		// Changeset() in after() - also empty, so they match
		// 		{Command: "git status --short --porcelain -z", Output: ""},
		// 		{Command: "git stash list", Output: "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
		// 		{Command: "git stash drop --quiet 1", Output: ""},
		// 	},
		// },
		// "failOnChanges=true with changeset error in after": {
		// 	stashUnstagedChanges: false,
		// 	failOnChanges:        true,
		// 	commands: []cmdtest.Out{
		// 		// Changeset() error in before() - empty output, so changesetBefore is empty
		// 		{Command: "git status --short --porcelain -z", Output: ""},
		// 		// Changeset() error in after() - empty output, so changesetAfter is empty
		// 		{Command: "git status --short --porcelain -z", Output: ""},
		// 	},
		// },
		"failOnChanges=true failOnChangesDiff=true with no changed files": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			failOnChangesDiff:    true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
				{Command: "git status --short --porcelain -z", Output: ""},
			},
		},
		// "stashUnstagedChanges=true with drop stash error": {
		// 	stashUnstagedChanges: true,
		// 	failOnChanges:        false,
		// 	commands: []cmdtest.Out{
		// 		{Command: "git status --short --porcelain -z", Output: "AM file1\x00"},
		// 		{Command: "git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
		// 			filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
		// 			" -- file1", Output: ""},
		// 		{Command: "git stash create", Output: "<stash-hash>"},
		// 		{Command: "git stash store --quiet --message lefthook auto backup <stash-hash>", Output: ""},
		// 		{Command: "git checkout --force -- file1", Output: ""},
		// 		// RestoreUnstaged succeeds (patch file will be created)
		// 		{Command: "git stash list", Output: ""}, // Empty list - stash not found, drop will fail silently
		// 	},
		// },
		"failOnChanges=true with deleted file in changeset": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: " M file1\x00"},
				{Command: "git hash-object -- file1", Output: "hash1\n"},
				{Command: "git status --short --porcelain -z", Output: "D  file1\x00"},
				// file1 was deleted, so it's in changesetAfter but marked as "deleted"
			},
			err: &FailOnChangesError{[]string{"file1"}},
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
			if tt.err != nil {
				assert.ErrorAs(tt.err, &err)
			} else {
				assert.NoError(err)
			}

			assert.Equal(true, beenCalled)
		})
	}
}
