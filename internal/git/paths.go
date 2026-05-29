package git

import (
	"path/filepath"
	"strings"
)

var cmdPaths = []string{
	"git", "rev-parse", "--path-format=absolute",
	"--show-toplevel",
	"--git-path", "hooks",
	"--git-path", "info",
	"--git-dir",
}

type GitPaths struct {
	RootPath  string
	HooksPath string
	InfoPath  string
	GitPath   string
}

func Paths(commander *Commander) (*GitPaths, error) {
	paths, err := commander.Cmd(cmdPaths)
	if err != nil {
		return nil, err
	}

	pathsSplit := strings.Split(paths, "\n")

	return &GitPaths{
		RootPath:  pathsSplit[0],
		HooksPath: pathsSplit[1],
		InfoPath:  filepath.Clean(pathsSplit[2]),
		GitPath:   pathsSplit[3],
	}, nil
}
