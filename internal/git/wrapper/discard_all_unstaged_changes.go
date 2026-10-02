package wrapper

var cmdDiscardAllUnstagedChanges = []string{"git", "checkout", "."}

func (w *Wrapper) DiscardAllUnstagedChanges() error {
	_, err := w.cmd.cmd(cmdDiscardAllUnstagedChanges)
	return err
}
