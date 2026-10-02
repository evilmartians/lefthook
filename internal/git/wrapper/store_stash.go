package wrapper

const stashMessage = "lefthook auto backup"

var cmdCreateStash = []string{"git", "stash", "create"}

func (w *Wrapper) StoreStash() error {
	stashHash, err := w.cmd.cmd(cmdCreateStash)
	if err != nil {
		return err
	}

	_, err = w.cmd.cmd([]string{
		"git",
		"stash",
		"store",
		"--quiet",
		"--message",
		stashMessage,
		stashHash,
	})
	return err
}
