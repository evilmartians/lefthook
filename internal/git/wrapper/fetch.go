package wrapper

import "fmt"

func (w *Wrapper) Fetch(ref, root string) error {
	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := w.cmd.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()

	_, err := git.cmd([]string{
		"git", "-C", root, "fetch", "--quiet", "--depth", "1",
		"origin", "--", ref,
	})
	if err != nil {
		return err
	}

	// A commit SHA can only be fetched from a remote that allows fetching
	// arbitrary objects. Make sure the fetched object is actually a commit
	// before checking it out.
	if isCommitSHA(ref) {
		_, err = git.cmd([]string{
			"git", "-C", root, "rev-parse", "--verify", "--quiet",
			ref + "^{commit}",
		})
		if err != nil {
			return fmt.Errorf("commit %s not found in the remote repository: %w", ref, err)
		}
	}

	_, err = git.cmd([]string{
		"git", "-C", root, "checkout", "FETCH_HEAD",
	})

	return err
}
