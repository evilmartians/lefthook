package run

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/run/controller"
	"github.com/evilmartians/lefthook/v2/internal/run/result"
)

var ErrFailOnChanges = controller.ErrFailOnChanges

type Options = controller.Options

func Run(ctx context.Context, hook *config.Hook, repo *git.Repository, opts Options) ([]result.Result, error) {
	return controller.NewController(repo).RunHook(ctx, opts, hook)
}
