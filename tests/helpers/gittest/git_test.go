package gittest

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func TestBuilder(t *testing.T) {
	fs := afero.NewMemMapFs()
	cmd := system.Cmd
	repo := NewRepositoryBuilder().Root("root").Fs(fs).Git(cmd).Build()

	assert := assert.New(t)
	assert.Equal("root", repo.RootPath)
	assert.Equal(filepath.Join("root", ".git"), repo.GitPath)
	assert.Equal(filepath.Join("root", ".git", "info"), repo.InfoPath)
	assert.Equal(filepath.Join("root", ".git", "hooks"), repo.HooksPath)
	assert.Equal(git.NewExecutor(cmd), repo.Git)
	assert.Equal(fs, repo.Fs)
}

func TestGitPath(t *testing.T) {
	assert.Equal(t, filepath.Join("root", ".git"), GitPath("root"))
}
