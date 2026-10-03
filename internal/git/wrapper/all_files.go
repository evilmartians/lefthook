package wrapper

var cmdAllFiles = []string{"git", "ls-files", "--cached"}

func (w *Wrapper) AllFiles() ([]string, error) {
	return w.Files(cmdAllFiles)
}
