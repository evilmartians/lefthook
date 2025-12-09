package command

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/internal/templates"
)

const (
	EnvVerbose             = "LEFTHOOK_VERBOSE" // keep all output
	envNoColor             = "NO_COLOR"
	envForceColor          = "CLICOLOR_FORCE"
	hookFileMode           = 0o755
	oldHookPostfix         = ".old"
	hookContentFingerprint = "LEFTHOOK"
)

type Lefthook struct {
	fs   afero.Fs
	repo *git.Repository
}

// NewLefthook returns an instance of Lefthook.
func NewLefthook(verbose bool, colors string) (*Lefthook, error) {
	fs := afero.NewOsFs()

	if isEnvEnabled(EnvVerbose) {
		verbose = true
	}

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if colors == "auto" {
		if isEnvEnabled(envForceColor) {
			colors = "on"
		}

		if isEnvEnabled(envNoColor) {
			colors = "off"
		}
	}

	log.SetColors(colors)

	repo, err := git.NewRepository(fs, git.NewExecutor(system.Cmd))
	if err != nil {
		return nil, err
	}

	return &Lefthook{fs: fs, repo: repo}, nil
}

func (l *Lefthook) LoadConfig() (*config.Config, error) {
	return config.Load(l.fs, l.repo)
}

// Tests a file whether it is a lefthook-created file.
func (l *Lefthook) isLefthookFile(path string) bool {
	file, err := l.fs.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Warnf("Could not close %s: %s", file.Name(), cErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), hookContentFingerprint) {
			return true
		}
	}
	if err = scanner.Err(); err != nil {
		log.Warnf("Could not read %s: %s", file.Name(), err)
	}

	return false
}

// Removes the hook from hooks path, saving non-lefthook hooks with .old suffix.
func (l *Lefthook) cleanHook(hook string, force bool) error {
	hookPath := filepath.Join(l.repo.HooksPath, hook)
	exists, err := afero.Exists(l.fs, hookPath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// Just remove lefthook hook
	if l.isLefthookFile(hookPath) {
		return l.fs.Remove(hookPath)
	}

	// Check if .old file already exists before renaming.
	exists, err = afero.Exists(l.fs, hookPath+oldHookPostfix)
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

	err = l.fs.Rename(hookPath, hookPath+oldHookPostfix)
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
		l.fs, hookPath, templates.Hook(hook, args), hookFileMode,
	)
}

func isEnvEnabled(name string) bool {
	value := os.Getenv(name)
	if len(value) > 0 && value != "0" && value != "false" {
		return true
	}

	return false
}

func ShellCompleteHookNames() {
	l, err := NewLefthook(false, "off")
	if err != nil {
		return
	}

	cfg, err := l.LoadConfig()
	if err != nil {
		return
	}

	for hook := range cfg.Hooks {
		fmt.Println(hook) //nolint:forbidigo // undecorated stdout is a must
	}
}

func ShellCompleteFlags(cmd *cli.Command) {
	given := cmd.FlagNames()
flags:
	for _, f := range cmd.VisibleFlags() {
		toAdd := make([]string, 0, len(f.Names()))
		for _, fn := range f.Names() {
			// Exclude all aliases of a flag if any of them is already given
			if slices.Contains(given, fn) {
				continue flags
			}
			// Do not bother with single letter flags.
			// If the user knows what they're for, they can just write them (hit the letter instead of tab),
			// no need to clutter the output with them.
			if len(fn) != 1 {
				toAdd = append(toAdd, fn)
			}
		}
		for _, fn := range toAdd {
			fmt.Println("--" + fn) //nolint:forbidigo // undecorated stdout is a must
		}
	}
}
