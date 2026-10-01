package wrapper

var cmdGitVersion = []string{"git", "version"}

func (w *Wrapper) Version() (string, error) {
	w.cmd.cmd(cmdGitVersion)
}
