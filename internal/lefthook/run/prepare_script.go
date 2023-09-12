package run

import (
	"errors"
	"os"
	"strings"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

func (r *Runner) prepareScript(script *config.Script, path string, file os.FileInfo) (string, error) {
	if script.DoSkip(r.Repo.State()) {
		return "", errors.New("settings")
	}

	if intersect(r.Hook.ExcludeTags, script.Tags) {
		return "", errors.New("excluded tags")
	}

	// Skip non-regular files (dirs, symlinks, sockets, etc.)
	if !file.Mode().IsRegular() {
		log.Debugf("[lefthook] file %s is not a regular file, skipping", file.Name())
		return "", errors.New("not a regular file")
	}

	// Make sure file is executable
	if (file.Mode() & executableMask) == 0 {
		if err := r.Repo.Fs.Chmod(path, executableFileMode); err != nil {
			log.Errorf("Couldn't change file mode to make file executable: %s", err)
			r.fail(file.Name(), nil)
			return "", errors.New("system error")
		}
	}

	var args []string
	if len(script.Runner) > 0 {
		args = append(args, script.Runner)
	}

	args = append(args, path)
	args = append(args, r.GitArgs...)

	return strings.Join(args, " "), nil
}
