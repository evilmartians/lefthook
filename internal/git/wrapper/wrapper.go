package wrapper

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

const (
	unstagedPatchName    = "lefthook-unstaged.patch"
	unstagedAllPatchName = "lefthook-unstaged-all.patch"
)

var (
	cmdRemotes         = []string{"git", "branch", "--remotes"}
	reHeadBranch       = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)
	reOriginHeadBranch = regexp.MustCompile(`ref: refs/remotes/origin/(?P<name>.*)$`)
)

// Wrapper wraps the calls to Git.
type Wrapper struct {
	// File system layer
	fs afero.Fs

	// Logger
	logger *logger.Logger

	// Command executor with system-dependent adjustments
	cmd *Cmd

	// Cached .git/info/ dir path for storing temporary files
	infoPath string

	// Cached .git/ dir path for head branch resolving
	gitPath string

	// Cached Git branch of the current HEAD
	headBranch string
}

func New(fs afero.Fs, logger *logger.Logger) *Wrapper {
	cmd := NewCmd(system.Cmd, logger)

	return &Wrapper{
		fs:     fs,
		logger: logger,
		cmd:    cmd,
	}
}

func (w *Wrapper) Files(cmd []string) ([]string, error) {
	return w.FilesRelative(cmd, "")
}

func (w *Wrapper) FilesRelative(cmd []string, dir string) ([]string, error) {
	lines, err := w.cmd.cmdLinesRelative(cmd, dir)
	if err != nil {
		return nil, err
	}

	return slices.Collect(w.selectFiles(unquoted(trimmed(slices.Values(lines))))), nil
}

// resolveHeadBranch determines the upstream head branch.
func (w *Wrapper) resolveHeadBranch() string {
	if branch := w.readOriginHead(); len(branch) > 0 {
		return branch
	}

	branches, err := w.cmd.cmdLines(cmdRemotes)
	if err == nil {
		for _, branch := range branches {
			matches := reHeadBranch.FindStringSubmatch(branch)
			if matches == nil {
				continue
			}
			return matches[reHeadBranch.SubexpIndex("name")]
		}
	}

	return ""

}

func (w *Wrapper) readOriginHead() string {
	originHead := filepath.Join(w.gitPath, "refs", "remotes", "origin", "HEAD")
	if _, err := w.fs.Stat(originHead); os.IsNotExist(err) {
		return ""
	}

	file, err := w.fs.Open(originHead)
	if err != nil {
		return ""
	}
	defer func() {
		if err := file.Close(); err != nil {
			w.logger.Warnf("Could not close %s: %s", originHead, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	_ = scanner.Scan()
	match := reOriginHeadBranch.FindStringSubmatch(scanner.Text())
	if match == nil {
		return ""
	}

	// Return the remote-tracking ref (e.g. "origin/main"), not the bare
	// branch name: the bare name may not exist as a local branch (a clone
	// that never checked out the default branch has no local "main").
	return "origin/" + match[reOriginHeadBranch.SubexpIndex("name")]
}

func (w *Wrapper) unstagedDiffPath() string {
	return filepath.Join(w.infoPath, unstagedPatchName)
}

func (w *Wrapper) unstagedAllDiffPath() string {
	return filepath.Join(w.infoPath, unstagedAllPatchName)
}
