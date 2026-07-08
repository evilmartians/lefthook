package app

import (
	"context"

	"github.com/evilmartians/lefthook/v2/internal/app"
)

type RunArgs struct {
	NoTTY             bool
	AllFiles          bool
	FilesFromStdin    bool
	Force             bool
	NoAutoInstall     bool
	NoStageFixed      bool
	SkipLFS           bool
	Verbose           bool
	FailOnChanges     *bool
	FailOnChangesDiff *bool
	Hook              string
	Exclude           []string
	Files             []string
	RunOnlyCommands   []string
	RunOnlyJobs       []string
	RunOnlyTags       []string
	GitArgs           []string
}

func Run(ctx context.Context, app *app.App, args RunArgs) error {
	return nil
}
