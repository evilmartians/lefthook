package lefthook

import (
	"bufio"
	"regexp"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/git"
	"github.com/evilmartians/lefthook/pkg/log"
)

var lefthookContentRegexp = regexp.MustCompile("LEFTHOOK")

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
		repo git.Repository
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

// initRepo initializes default repository object, if it wasn't assigned before
func initRepo(lefthook *Lefthook) error {
	if lefthook.repo != nil {
		return nil
	}

	repo, err := git.NewRepository()
	if err != nil {
		return err
	}

	lefthook.repo = repo

	return nil
}

func (l Lefthook) isLefthookFile(path string) bool {
	file, err := l.fs.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if lefthookContentRegexp.MatchString(scanner.Text()) {
			return true
		}
	}

	return false
}
