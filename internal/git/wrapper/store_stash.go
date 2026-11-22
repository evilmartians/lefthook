package wrapper

// backupRefPrefix keeps backups out of the stash list, which all worktrees
// share, so a cleanup can only ever delete its own backup.
const backupRefPrefix = "refs/lefthook/backup/"

var cmdCreateStash = []string{"git", "stash", "create"}

func (w *Wrapper) StoreStash() error {
	stashHash, err := w.cmd.cmd(cmdCreateStash)
	if err != nil {
		return err
	}

	_, err = w.cmd.cmd([]string{
		"git",
		"update-ref",
		backupRefPrefix + stashHash,
		stashHash,
		"",
	})
	if err != nil {
		return err
	}

	w.cache.backupStashHash = stashHash

	return nil
}
