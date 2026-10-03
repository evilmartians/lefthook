package wrapper

func (w *Wrapper) Pull(root string) error {
	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := w.cmd.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()
	_, err := git.cmd([]string{"git", "-C", root, "pull", "--quiet"})
	return err
}
