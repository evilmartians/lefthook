package lefthook

import (
	"bufio"
	"regexp"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/git"
	"github.com/evilmartians/lefthook/pkg/log"
)

var lefthookContentRegexp = regexp.MustCompile("LEFTHOOK")

type Options struct {
	Fs                afero.Fs
	Verbose, NoColors bool

	// DEPRECATED
	Force, Aggressive bool
}

type Lefthook struct {
	// Since we need to support deprecated global options Force and Aggressive
	// we need to store these fields. After their removal we need just to copy fs.
	*Options

	repo git.Repository
}

// New returns an instance of Lefthook
func initialize(opts *Options) (*Lefthook, error) {
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.SetColors(!opts.NoColors)

	repo, err := git.NewRepository()
	if err != nil {
		return nil, err
	}

	return &Lefthook{Options: opts, repo: repo}, nil
}

func (l *Lefthook) isLefthookFile(path string) bool {
	file, err := l.Fs.Open(path)
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
