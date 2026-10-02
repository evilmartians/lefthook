package gittest

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/system"
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

func (b *RepositoryBuilder) BuildRepo() *git.Repo {
	repo := &git.Repo{
		Fs: b.fs,
		Paths: git.Paths{
			Root:  b.root,
			Git:   GitPath(b.root),
			Hooks: filepath.Join(GitPath(b.root), "hooks"),
			Info:  filepath.Join(GitPath(b.root), "info"),
		},
	}

	return repo
}

func GitPath(root string) string {
	return filepath.Join(root, ".git")
}
