package wrapper

const stashMessage = "lefthook auto backup"

var cmdCreateStash = []string{"git", "stash", "create"}

func (w *Wrapper) StoreStash() error {
	stashHash, err := w.cmd.cmd(cmdCreateStash)
	if err != nil {
		return err
	}

	message := stashMessage
	if w.cache.worktree {
		message += " - " + w.cache.projectDirname()
	}

	_, err = w.cmd.cmd([]string{
		"git",
		"stash",
		"store",
		"--quiet",
		"--message",
		message,
		stashHash,
	})
	if err != nil {
		return err
	}

	w.cache.backupStashHash = stashHash

	return nil
}
