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

func (s *GitService) SyncRemote(ctx context.Context, remote *config.Remote, force bool) error {
	return nil
}
