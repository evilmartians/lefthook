package wrapper

import "fmt"

func (w *Wrapper) SaveUnstagedDiff(files []string) error {
	_, err := w.cmd.batchedCmd(
		[]string{
			"git",
			"diff",
			"--binary",          // support binary files
			"--unified=0",       // do not add lines around diff for consistent behavior
			"--no-color",        // disable colors for consistent behavior
			"--no-ext-diff",     // disable external diff tools for consistent behavior
			"--src-prefix=a/",   // force prefix for consistent behavior
			"--dst-prefix=b/",   // force prefix for consistent behavior
			"--patch",           // output a patch that can be applied
			"--submodule=short", // always use the default short format for submodules
			"--output",
			w.unstagedDiffPath(),
			"--",
		}, files)

	if err != nil {
		return fmt.Errorf("failed to create a diff for files: %w", err)
	}

	_, err = w.cmd.cmd([]string{
		"git",
		"diff",
		"--binary",
		"--unified=0",
		"--no-color",
		"--no-ext-diff",
		"--src-prefix=a/",
		"--dst-prefix=b/",
		"--patch",
		"--submodule=short",
		"--output",
		w.unstagedAllDiffPath(),
		"--",
	})
	if err != nil {
		return fmt.Errorf("failed to create a diff for all files: %w", err)
	}

	return nil
}
