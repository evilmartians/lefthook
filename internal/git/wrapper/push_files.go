package wrapper

var cmdPushFilesBase = []string{
	"git", "diff", "--name-only", "HEAD", "@{push}",
}

var cmdPushFilesHead = []string{
	"git", "diff", "--name-only", // origin/main...HEAD
}

var cmdLsTreeFilesHead = []string{
	"git", "ls-tree", "-r", "--name-only", "HEAD",
}

// PushFiles returns the list of files changed between the local branch
// and the remote branch.
//
// The list of files will be the best effort of:
// 1. Trying to compare current HEAD with @{push} ref
// 2. Trying to compare current HEAD with origin/<current-branch>
// 3. Returning all files known to Git.
func (w *Wrapper) PushFiles() ([]string, error) {
	// Upstream can be unset, do not log the failure as an error
	lines, err := w.cmd.OnlyDebugLogs().cmdLines(cmdPushFilesBase)
	if err == nil { // ignoring error for best effort
		return w.existingFilepaths(lines), nil
	}

	if len(w.cache.headBranch) == 0 {
		w.cache.headBranch = w.resolveHeadBranch()
	}

	if len(w.cache.headBranch) != 0 {
		lines, err = w.cmd.OnlyDebugLogs().cmdLines(append(cmdPushFilesHead, w.cache.headBranch+"...HEAD", "--"))
		if err == nil { // ignoring error for best effort
			return w.existingFilepaths(lines), nil
		}
	}

	// Nothing has been pushed yet or upstream is not set
	return w.Files(cmdLsTreeFilesHead)
}
