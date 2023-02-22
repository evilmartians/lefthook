package runner

import (
	"errors"
	"os"
	"strings"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

func (r *Runner) prepareScript(script *config.Script, path string, file os.FileInfo) ([]string, error) {
	if script.Skip != nil && script.DoSkip(r.Repo.State()) {
		return nil, errors.New("settings")
	}

	if intersect(r.Hook.ExcludeTags, script.Tags) {
		return nil, errors.New("excluded tags")
	}

	// Skip non-regular files (dirs, symlinks, sockets, etc.)
	if !file.Mode().IsRegular() {
		log.Debugf("[lefthook] file %s is not a regular file, skipping", file.Name())
		return nil, errors.New("not a regular file")
	}

	// Make sure file is executable
	if (file.Mode() & executableMask) == 0 {
		if err := r.Fs.Chmod(path, executableFileMode); err != nil {
			log.Errorf("Couldn't change file mode to make file executable: %s", err)
			r.fail(file.Name(), "")
			return nil, errors.New("system error")
		}
	}

	var args []string
	if len(script.Runner) > 0 {
		args = strings.Split(script.Runner, " ")
	}

	args = append(args, path)
	args = append(args, r.GitArgs...)

	return args, nil
}
