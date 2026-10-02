package wrapper

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

	_, err = git.cmd([]string{
		"git", "-C", root, "checkout", "FETCH_HEAD",
	})

	return err
}
