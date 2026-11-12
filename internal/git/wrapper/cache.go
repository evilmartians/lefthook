package wrapper

import "path/filepath"

type cache struct {
	// .git/info/ dir path for storing temporary files
	infoPath string

	// .git/ dir path for head branch resolving
	gitPath string

	// Git branch of the current HEAD
	headBranch string

	// Filepath of the patch with unstaged changes
	unstagedPatchPath string

	// Filepath of the patch with all unstaged changes (including untracked files)
	unstagedAllPatchPath string

	// Git stash hash for backup of unstaged changes
	backupStashHash string

	// When inside a git worktree (a dir for the branch)
	worktree bool
}

func (c *cache) projectDirname() string {
	if c.worktree {
		return filepath.Base(c.gitPath)
	}

	return filepath.Base(filepath.Dir(c.gitPath))
}
