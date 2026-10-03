package wrapper

var cmdDiscardUnstagedChanges = []string{"git", "checkout", "--force", "--"}

func (w *Wrapper) DiscardUnstagedChanges(files []string) error {
	_, err := w.cmd.batchedCmd(cmdDiscardUnstagedChanges, files)
	return err
}
