package wrapper

func (w *Wrapper) GlobalHooksPath() string {
	res, _ := w.cmd.cmd([]string{"git", "config", "--global", "core.hooksPath"})
	return res
}
