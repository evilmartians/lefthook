package wrapper

import "github.com/spf13/afero"

func (w *Wrapper) UnstagedDiffApplicable() bool {
	return w.diffApplicable(w.unstagedPatchPath())
}

func (w *Wrapper) diffApplicable(diffPath string) bool {
	if ok, _ := afero.Exists(w.fs, diffPath); !ok {
		return true
	}

	stat, err := w.fs.Stat(diffPath)
	if err != nil {
		return true
	}

	if stat.Size() == 0 {
		return true
	}

	_, err = w.cmd.cmd([]string{
		"git",
		"apply",
		"-v",
		"--whitespace=nowarn",
		"--recount",
		"--unidiff-zero",
		"--check",
		"--",
		diffPath,
	})

	return err == nil
}
