package lefthook

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/templates"
)

const hookFileMode = 0o755

var lefthookContentRegexp = regexp.MustCompile("LEFTHOOK")

type Options struct {
	Fs                afero.Fs
	Verbose, NoColors bool

	// DEPRECATED. Will be removed in 1.3.0.
	Force, Aggressive bool
}

type Lefthook struct {
	// Since we need to support deprecated global options Force and Aggressive
	// we need to store these fields. After their removal we need just to copy fs.
	*Options

	repo *git.Repository
}

// New returns an instance of Lefthook.
func initialize(opts *Options) (*Lefthook, error) {
	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.SetColors(!opts.NoColors)

	repo, err := git.NewRepository(opts.Fs, git.NewOsExec())
	if err != nil {
		return nil, err
	}

	return &Lefthook{Options: opts, repo: repo}, nil
}

// Tests a file whether it is a lefthook-created file.
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

// Removes the hook from hooks path, saving non-lefthook hooks with .old suffix.
func (l *Lefthook) cleanHook(hook string, force bool) error {
	hookPath := filepath.Join(l.repo.HooksPath, hook)
	exists, err := afero.Exists(l.Fs, hookPath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// Just remove lefthook hook
	if l.isLefthookFile(hookPath) {
		if err = l.Fs.Remove(hookPath); err != nil {
			return err
		}

		return nil
	}

	// Check if .old file already exists before renaming.
	exists, err = afero.Exists(l.Fs, hookPath+".old")
	if err != nil {
		return err
	}
	if exists {
		if force {
			log.Infof("File %s.old already exists, overwriting\n", hook)
		} else {
			log.Errorf("Can't rename %s to %s.old - file already exists\n", hook, hook)
			return fmt.Errorf("file %s.old already exists", hook)
		}
	}

	err = l.Fs.Rename(hookPath, hookPath+".old")
	if err != nil {
		return err
	}

	log.Infof("Renamed %s to %s.old\n", hookPath, hookPath)
	return nil
}

// Creates a hook file using hook template.
func (l *Lefthook) addHook(hook, rc string) error {
	hookPath := filepath.Join(l.repo.HooksPath, hook)
	return afero.WriteFile(
		l.Fs, hookPath, templates.Hook(hook, rc), hookFileMode,
	)
}
