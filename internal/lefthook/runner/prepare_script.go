package runner

import (
	"os"
	"strings"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

func (r *Runner) prepareScript(script *config.Script, path string, file os.FileInfo) (string, error) {
	if script.DoSkip(r.Repo.State()) {
		return "", &skipError{"settings"}
	}

	if intersect(r.Hook.ExcludeTags, script.Tags) {
		return "", &skipError{"excluded tags"}
	}

	// Skip non-regular files (dirs, symlinks, sockets, etc.)
	if !file.Mode().IsRegular() {
		log.Debugf("[lefthook] file %s is not a regular file, skipping", file.Name())
		return "", &skipError{"not a regular file"}
	}

	// Make sure file is executable
	if (file.Mode() & executableMask) == 0 {
		if err := r.Repo.Fs.Chmod(path, executableFileMode); err != nil {
			log.Errorf("Couldn't change file mode to make file executable: %s", err)
			return "", err
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
