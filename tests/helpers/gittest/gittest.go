package gittest

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

type RepositoryBuilder struct {
	root string
	cmd  system.Command
	fs   afero.Fs
}

func NewRepositoryBuilder() *RepositoryBuilder {
	return &RepositoryBuilder{}
}

func (b *RepositoryBuilder) Root(root string) *RepositoryBuilder {
	b.root = root
	return b
}

func (b *RepositoryBuilder) Cmd(cmd system.Command) *RepositoryBuilder {
	b.cmd = cmd
	return b
}

func (b *RepositoryBuilder) Fs(fs afero.Fs) *RepositoryBuilder {
	b.fs = fs
	return b
}

func (b *RepositoryBuilder) Build() *git.Repo {
	logger := loggertest.New()
	cmd := &pathsCmd{
		next: b.cmd,
		output: strings.Join([]string{
			b.root,
			filepath.Join(GitPath(b.root), "hooks"),
			filepath.Join(GitPath(b.root), "info"),
			GitPath(b.root),
			GitPath(b.root),
		}, "\n"),
	}
	w := wrapper.New(b.fs, cmd, logger)

	// Initialize the wrapper paths without passing the call to the test command
	paths, err := w.Paths()
	if err != nil {
		panic(err)
	}

	return git.NewRepo(b.fs, logger, w, paths)
}

func GitPath(root string) string {
	return filepath.Join(root, ".git")
}

// pathsCmd answers the first `git rev-parse` call with the builder paths
// and passes all other calls to the next command.
type pathsCmd struct {
	next   system.Command
	output string
	done   bool
}

func (c *pathsCmd) WithoutEnvs(envs ...string) system.Command {
	if c.done {
		return c.next.WithoutEnvs(envs...)
	}

	return c
}

func (c *pathsCmd) Run(command []string, root string, in io.Reader, out io.Writer, errOut io.Writer) error {
	if !c.done && len(command) > 1 && command[1] == "rev-parse" {
		c.done = true
		_, err := out.Write([]byte(c.output))
		return err
	}

	return c.next.Run(command, root, in, out, errOut)
}
