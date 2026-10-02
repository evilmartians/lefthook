package wrapper

var (
	cmdDiffColored  = []string{"git", "diff", "--color", "--"}
	cmdDiffNoColors = []string{"git", "diff", "--"}
)

func (w *Wrapper) Diff(files []string, colors bool) (string, error) {
	if colors {
		return w.cmd.batchedCmd(cmdDiffColored, files)
	}

	return w.cmd.batchedCmd(cmdDiffNoColors, files)
}
