package lefthook

import (
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/git"
	"github.com/evilmartians/lefthook/pkg/log"
)

type (
	Options struct {
		Fs                afero.Fs
		Verbose, NoColors bool

		// DEPRECATED
		Force, Aggressive bool
	}

	Lefthook struct {
		opts *Options
		fs   afero.Fs
		repo *git.Repository
	}
)

// New returns an instance of Lefthook
func New(opts *Options) Lefthook {
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// TODO: check NoColors

	return Lefthook{opts: opts, fs: opts.Fs}
}

func initRepo(lefthook *Lefthook) error {
	repo, err := git.NewRepository()
	if err != nil {
		return err
	}

	lefthook.repo = repo

	return nil
}

func (l Lefthook) isLefthookFile(path string) bool {
	file, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return false
	}

	return strings.Contains(string(file), "LEFTHOOK")
}
