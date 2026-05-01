package app

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
)

type GitService struct {
	repo   *git.Repo
	logger *logger.Logger
}

// SyncRemote clones or pulls the configured remote so its hooks/extends are
// available locally. It mirrors the old `Lefthook.Install` behavior: always
// fetch when called.
func (s *GitService) SyncRemote(ctx context.Context, remote *config.Remote, force bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.repo.SyncRemote(remote.GitURL, remote.Ref, force)
}
