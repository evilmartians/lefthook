package jobs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/internal/log"
)

const (
	executableFileMode os.FileMode = 0o751
	executableMask     os.FileMode = 0o111
)

type scriptNotExistsError struct {
	scriptPath string
}

func (s scriptNotExistsError) Error() string {
	return fmt.Sprintf("script does not exist: %s", s.scriptPath)
}

func buildScript(params *Params, settings *Settings) (*Job, error) {
	if err := params.validateScript(); err != nil {
		return nil, err
	}

	var scriptExists bool
	execs := make([]string, 0)
	for _, sourceDir := range settings.SourceDirs {
		scriptPath := filepath.Join(sourceDir, settings.HookName, params.Script)
		fileInfo, err := settings.Repo.Fs.Stat(scriptPath)
		if os.IsNotExist(err) {
			log.Debugf("[lefthook] script doesn't exist: %s", scriptPath)
			continue
		}
		if err != nil {
			log.Errorf("Failed to get info about a script: %s", params.Script)
			return nil, err
		}

		scriptExists = true

		if !fileInfo.Mode().IsRegular() {
			log.Debugf("[lefthook] script '%s' is not a regular file, skipping", scriptPath)
			return nil, &SkipError{"not a regular file"}
		}

		// Make sure file is executable
		if (fileInfo.Mode() & executableMask) == 0 {
			if err := settings.Repo.Fs.Chmod(scriptPath, executableFileMode); err != nil {
				log.Errorf("Couldn't change file mode to make file executable: %s", err)
				return nil, err
			}
		}

		var args []string
		if len(params.Runner) > 0 {
			args = append(args, params.Runner)
		}

		args = append(args, shellescape.Quote(scriptPath))
		args = append(args, settings.GitArgs...)

		execs = append(execs, strings.Join(args, " "))
	}

	if !scriptExists {
		return nil, scriptNotExistsError{params.Script}
	}

	return &Job{
		Execs: execs,
		Files: []string{},
	}, nil
}
