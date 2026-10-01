package wrapper

var cmdStagedFiles = []string{
	"git", "diff", "--name-only", "--cached", "--diff-filter=ACMR",
}

func (w *Wrapper) StagedFiles() ([]string, error) {
	return w.Files(cmdStagedFiles)
}
