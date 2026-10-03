package wrapper

var cmdStageFiles = []string{"git", "add", "--force", "--"}

func (w *Wrapper) StageFiles(files []string) error {
	_, err := w.cmd.batchedCmd(cmdStageFiles, files)

	return err
}
