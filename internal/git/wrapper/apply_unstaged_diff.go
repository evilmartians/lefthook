package wrapper

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

// ErrNoUnstagedDiff is returned when there is no saved diff to apply.
var ErrNoUnstagedDiff = errors.New("no saved unstaged diff")

// ApplyUnstagedDiff applies the saved diff and removes the diff files.
// It returns ErrNoUnstagedDiff when no diff was saved.
func (w *Wrapper) ApplyUnstagedDiff(all bool) error {
	var diffPath string
	if all {
		diffPath = w.unstagedAllDiffPath()
	} else {
		diffPath = w.unstagedDiffPath()
	}

	exists, err := afero.Exists(w.fs, diffPath)
	if err != nil {
		return fmt.Errorf("failed to inspect the patch %s: %w", diffPath, err)
	}
	if !exists {
		return ErrNoUnstagedDiff
	}

	stat, err := w.fs.Stat(diffPath)
	if err != nil {
		return err
	}

	if stat.Size() > 0 {
		_, err = w.cmd.cmd([]string{
			"git",
			"apply",
			"-v",
			"--whitespace=nowarn",
			"--recount",
			"--unidiff-zero",
			"--",
			diffPath,
		})
		if err != nil {
			return fmt.Errorf("failed to apply the patch %s: %w", diffPath, err)
		}
	}

	if err = w.removeUnstagedDiff(w.unstagedDiffPath()); err != nil {
		return fmt.Errorf("failed to remove the patch %s: %w", w.unstagedDiffPath(), err)
	}

	if err = w.removeUnstagedDiff(w.unstagedAllDiffPath()); err != nil {
		return fmt.Errorf("failed to remove the patch %s: %w", w.unstagedAllDiffPath(), err)
	}

	return nil
}

func (w *Wrapper) removeUnstagedDiff(path string) error {
	exists, existsErr := afero.Exists(w.fs, path)
	if existsErr != nil {
		return fmt.Errorf("failed to inspect the patch %s: %w", path, existsErr)
	}
	if !exists {
		return nil
	}

	return w.fs.Remove(path)
}
