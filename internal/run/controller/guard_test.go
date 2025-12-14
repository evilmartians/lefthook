package controller

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

type guardCmd struct {
	commands [][2]string
}

func (c *guardCmd) WithoutEnvs(...string) system.Command {
	return c
}

func (c *guardCmd) Run(cmds []string, _root string, _in io.Reader, out io.Writer, _errOut io.Writer) error {
	if len(c.commands) == 0 {
		panic("expected: no command, called: " + "'" + strings.Join(cmds, " ") + "'")
	}

	commandL := c.commands[0][0]
	commandR := c.commands[0][1]

	if commandL != strings.Join(cmds, " ") {
		panic("expected: '" + commandL + "', called: " + "'" + strings.Join(cmds, " ") + "'")
	}

	_, _ = out.Write([]byte(commandR))
	c.commands = c.commands[1:]
	return nil
}

func Test_guard_wrap(t *testing.T) {
	for name, tt := range map[string]struct {
		stashUnstagedChanges bool
		failOnChanges        bool
		commands             [][2]string
		err                  error
	}{
		"just call": {
			stashUnstagedChanges: false,
			failOnChanges:        false,
			commands:             [][2]string{},
		},
		"failOnChanges=true no files": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: [][2]string{
				{"git status --short --porcelain -z", ""},
				{"git status --short --porcelain -z", ""},
			},
		},
		"failOnChanges=true no fail": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: [][2]string{
				{"git status --short --porcelain -z", " M file1\x00 M file2\x00"},
				{"git hash-object -- file1 file2", "0\n1\n"},
				{"git status --short --porcelain -z", " M file1\x00 M file2\x00"},
				{"git hash-object -- file1 file2", "0\n1\n"},
			},
		},
		"failOnChanges=true fail with changeset different": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: [][2]string{
				{"git status --short --porcelain -z", " M file1\x00 M file2\x00"},
				{"git hash-object -- file1 file2", "0\n1\n"},
				{"git status --short --porcelain -z", " M file1\x00 M file2\x00"},
				{"git hash-object -- file1 file2", "2\n3\n"},
			},
			err: ErrFailOnChanges,
		},
		"failOnChanges=true fail with extra files": {
			stashUnstagedChanges: false,
			failOnChanges:        true,
			commands: [][2]string{
				{"git status --short --porcelain -z", ""},
				{"git status --short --porcelain -z", " M file1\x00 M file2\x00"},
				{"git hash-object -- file1 file2", "0\n1\n"},
			},
			err: ErrFailOnChanges,
		},
		"stashUnstagedChanges=true no files": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: [][2]string{
				{"git status --short --porcelain -z", ""},
			},
		},
		"stashUnstagedChanges=true no unstaged": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: [][2]string{
				{"git status --short --porcelain -z", "M  file1\x00M  file2\x00M  file3\x00"},
			},
		},
		"stashUnstagedChanges=true with partially staged": {
			stashUnstagedChanges: true,
			failOnChanges:        false,
			commands: [][2]string{
				{"git status --short --porcelain -z", "AM file1\x00 M file2\x00 A file3\x00"},
				{"git diff --binary --unified=0 --no-color --no-ext-diff --src-prefix=a/ --dst-prefix=b/ --patch --submodule=short --output " +
					filepath.Join("root", ".git", "info", "lefthook-unstaged.patch") +
					" -- file1", ""},
				{"git stash create", "<stash-hash>"},
				{"git stash store --quiet --message lefthook auto backup <stash-hash>", ""},
				{"git checkout --force -- file1", ""},
				{"git stash list", "0: my stash\n1: lefthook auto backup\n2: my second stash\n"},
				{"git stash drop --quiet 1", ""},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			fs := afero.NewMemMapFs()
			repo := gittest.NewRepositoryBuilder().Git(&guardCmd{tt.commands}).Fs(fs).Root("root").Build()
			repo.Setup()
			g := newGuard(repo, tt.stashUnstagedChanges, tt.failOnChanges)

			var beenCalled bool
			err := g.wrap(func() { beenCalled = true })
			assert.Equal(tt.err, err)
			assert.Equal(true, beenCalled)
		})
	}
}
