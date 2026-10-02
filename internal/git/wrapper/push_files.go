package wrapper

var cmdPushFilesBase = []string{
	"git", "diff", "--name-only", "HEAD", "@{push}",
}

var cmdPushFilesHead = []string{
	"git", "diff", "--name-only", "HEAD",
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
	pushFiles, err := w.Files(cmdPushFilesBase)
	if err == nil {
		return pushFiles, nil
	}

	if len(w.headBranch) == 0 {
		w.headBranch = w.resolveHeadBranch()
	}

	if len(w.headBranch) != 0 {
		return w.Files(append(cmdPushFilesHead, w.headBranch, "--"))
	}

	// Nothing has been pushed yet or upstream is not set
	return w.Files(cmdLsTreeFilesHead)
}
