package run

import (
	"context"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/run/controller"
	"github.com/evilmartians/lefthook/internal/run/result"
)

var ErrFailOnChanges = controller.ErrFailOnChanges

type Options = controller.Options

func Run(ctx context.Context, hook *config.Hook, repo *git.Repository, opts Options) ([]result.Result, error) {
	return controller.NewController(repo).RunHook(ctx, opts, hook)
}
