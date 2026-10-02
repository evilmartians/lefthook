package wrapper

var cmdGitVersion = []string{"git", "version"}

func (w *Wrapper) Version() (string, error) {
	return w.cmd.cmd(cmdGitVersion)
}
