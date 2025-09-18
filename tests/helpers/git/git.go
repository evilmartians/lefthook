package git

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
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

func (b *RepositoryBuilder) Git(cmd system.Command) *RepositoryBuilder {
	b.cmd = cmd
	return b
}

func (b *RepositoryBuilder) Fs(fs afero.Fs) *RepositoryBuilder {
	b.fs = fs
	return b
}

func (b *RepositoryBuilder) Build() *git.Repository {
	return &git.Repository{
		Fs:        b.fs,
		Git:       git.NewExecutor(b.cmd),
		RootPath:  b.root,
		GitPath:   GitPath(b.root),
		HooksPath: filepath.Join(GitPath(b.root), "hooks"),
		InfoPath:  filepath.Join(GitPath(b.root), "info"),
	}
}

func GitPath(root string) string {
	return filepath.Join(root, ".git")
}
