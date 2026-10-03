package wrapper

func (w *Wrapper) LocalHooksPath() string {
	res, _ := w.cmd.cmd([]string{"git", "config", "--local", "core.hooksPath"})
	return res
}
