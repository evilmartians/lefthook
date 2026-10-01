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

type PathsResult struct {
	Root  string
	Hooks string
	Info  string
	Git   string
}

func (w *Wrapper) Paths() (*PathsResult, error) {
	paths, err := w.cmd.cmdLines(cmdPaths)
	if err != nil {
		return nil, err
	}

	// Refresh the current paths
	w.cmd.root = paths[0]
	w.infoPath = filepath.Clean(paths[2])
	w.gitPath = paths[3]

	return &PathsResult{
		Root:  paths[0],
		Hooks: paths[1],
		Info:  w.infoPath,
		Git:   w.gitPath,
	}, nil
}
