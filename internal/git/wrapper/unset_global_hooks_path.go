package wrapper

func (w *Wrapper) UnsetGlobalHooksPath() error {
	_, err := w.cmd.cmd([]string{"git", "config", "--global", "--unset-all", "core.hooksPath"})
	return err
}
