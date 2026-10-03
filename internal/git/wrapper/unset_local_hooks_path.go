package wrapper

func (w *Wrapper) UnsetLocalHooksPath() error {
	_, err := w.cmd.cmd([]string{"git", "config", "--local", "--unset-all", "core.hooksPath"})
	return err
}
