package lefthook

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
	"github.com/evilmartians/lefthook/internal/templates"
)

const (
	EnvVerbose             = "LEFTHOOK_VERBOSE" // keep all output
	envNoColor             = "NO_COLOR"
	envForceColor          = "CLICOLOR_FORCE"
	hookFileMode           = 0o755
	oldHookPostfix         = ".old"
	hookContentFingerprint = "LEFTHOOK"
)

type Options struct {
	Fs                afero.Fs
	Verbose, NoColors bool
	Colors            string

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
	if isEnvEnabled(EnvVerbose) {
		opts.Verbose = true
	}

	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if opts.Colors == "auto" {
		if isEnvEnabled(envForceColor) {
			opts.Colors = "on"
		}

		if isEnvEnabled(envNoColor) {
			opts.Colors = "off"
		}

		// DEPRECATED: Will be removed with a --no-colors option
		if opts.NoColors {
			opts.Colors = "off"
		}
	}

	log.SetColors(opts.Colors)

	repo, err := git.NewRepository(opts.Fs, git.NewExecutor(system.Cmd))
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
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Warnf("Could not close %s: %s", file.Name(), cErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), hookContentFingerprint) {
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
		return l.Fs.Remove(hookPath)
	}

	// Check if .old file already exists before renaming.
	exists, err = afero.Exists(l.Fs, hookPath+oldHookPostfix)
	if err != nil {
		return err
	}
	if exists {
		if force {
			log.Infof("\nFile %s.old already exists, overwriting\n", hook)
		} else {
			return fmt.Errorf("can't rename %s to %s.old - file already exists", hook, hook)
		}
	}

	err = l.Fs.Rename(hookPath, hookPath+oldHookPostfix)
	if err != nil {
		return err
	}

	log.Infof("Renamed %s to %s.old\n", hookPath, hookPath)
	return nil
}

// Creates a hook file using hook template.
func (l *Lefthook) addHook(hook string, args templates.Args) error {
	hookPath := filepath.Join(l.repo.HooksPath, hook)
	return afero.WriteFile(
		l.Fs, hookPath, templates.Hook(hook, args), hookFileMode,
	)
}

func isEnvEnabled(name string) bool {
	value := os.Getenv(name)
	if len(value) > 0 && value != "0" && value != "false" {
		return true
	}

	return false
}
