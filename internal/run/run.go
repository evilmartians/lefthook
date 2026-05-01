package run

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/run/controller"
	"github.com/evilmartians/lefthook/v2/internal/run/result"
)

// FailOnChangesError is a special error that fails the hook if any project file was changed.
//
// Exported here to be handled separately on the caller side.
type FailOnChangesError = controller.FailOnChangesError

// Options contain hook arguments and special execution settings.
type Options = controller.Options

// Run executes the hook.
func Run(
	ctx context.Context,
	hook *config.Hook,
	repo *git.Repository,
	opts Options,
) ([]result.Result, error) {
	return controller.NewController(repo).RunHook(ctx, opts, hook)
}
