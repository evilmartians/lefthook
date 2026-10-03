package wrapper

import (
	"path/filepath"
)

var cmdPaths = []string{
	"git", "rev-parse", "--path-format=absolute",
	"--show-toplevel",
	"--git-path", "hooks",
	"--git-path", "info",
	"--git-dir",
}

type Paths struct {
	// Project root path
	Root string

	// .git/hooks/ dir path
	Hooks string

	// .git/ dir path
	Git string

	// .git/info/ dir path
	Info string
}

func (w *Wrapper) Paths() (*Paths, error) {
	paths, err := w.cmd.cmdLines(cmdPaths)
	if err != nil {
		return nil, err
	}

	root := paths[0]
	// `git rev-parse --show-toplevel` reports the current directory instead
	// of the worktree root when GIT_DIR is set explicitly and the current
	// directory is elsewhere. This happens with tools like `uv run`, which
	// execute with GIT_DIR pointing at the real repository while running
	// from a package-manager cache directory. `--git-dir` still reports
	// the true location, so fall back to its parent directory when the
	// reported root does not actually contain it.
	if gitDir := paths[3]; !w.gitDirMatchesRoot(root, gitDir) &&
		filepath.Base(gitDir) == ".git" {
		root = filepath.Dir(gitDir)
	}

	// Refresh the current paths
	w.cmd.root = root
	w.infoPath = filepath.Clean(paths[2])
	w.gitPath = paths[3]

	return &Paths{
		Root:  root,
		Hooks: paths[1],
		Info:  w.infoPath,
		Git:   w.gitPath,
	}, nil
}

// gitDirMatchesRoot reports whether root plausibly contains gitDir: either
// it is the conventional <root>/.git directory, or <root>/.git exists as a
// gitlink file (worktree, submodule).
func (w *Wrapper) gitDirMatchesRoot(root, gitDir string) bool {
	if gitDir == filepath.Join(root, ".git") {
		return true
	}
	if _, err := w.fs.Stat(filepath.Join(root, ".git")); err == nil {
		return true
	}
	return false
}
