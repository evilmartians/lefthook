package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/command/replacer"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
	"github.com/evilmartians/lefthook/v2/internal/system"
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

func (b *Builder) buildScript(params *JobParams) ([]string, []string, error) {
	if err := params.validateScript(); err != nil {
		return nil, nil, err
	}

	var replacer replacer.Replacer
	if len(params.Args) > 0 {
		replacer = b.buildReplacer(params)
		// TODO(mrexox): Duplicate filter creation is not fancy
		filter := filter.New(b.git.Fs, filter.Params{
			Glob:         params.Glob,
			ExcludeFiles: params.ExcludeFiles,
			Root:         params.Root,
			FileTypes:    params.FileTypes,
			GlobMatcher:  b.opts.GlobMatcher,
		})
		err := replacer.Discover(params.Args, filter)
		if err != nil {
			return nil, nil, err
		}
	}

	var scriptExists bool
	execs := make([]string, 0)
	for _, sourceDir := range b.opts.SourceDirs {
		scriptPath := filepath.Join(sourceDir, b.opts.HookName, params.Script)
		fileInfo, err := b.git.Fs.Stat(scriptPath)
		if os.IsNotExist(err) {
			log.Debugf("[lefthook] script doesn't exist: %s", scriptPath)
			continue
		}
		if err != nil {
			log.Errorf("Failed to get info about a script: %s", params.Script)
			return nil, nil, err
		}

		scriptExists = true

		if !fileInfo.Mode().IsRegular() {
			log.Debugf("[lefthook] script '%s' is not a regular file, skipping", scriptPath)
			return nil, nil, SkipError{"not a regular file"}
		}

		// Make sure file is executable
		if (fileInfo.Mode() & executableMask) == 0 {
			if err := b.git.Fs.Chmod(scriptPath, executableFileMode); err != nil {
				log.Errorf("Couldn't change file mode to make file executable: %s", err)
				return nil, nil, err
			}
		}

		var args []string
		if len(params.Runner) > 0 {
			args = append(args, params.Runner)
		}

		args = append(args, shellescape.Quote(scriptPath))
		if len(params.Args) > 0 {
			args = append(args, params.Args)
			command := strings.Join(args, " ")
			commands, _ := replacer.ReplaceAndSplit(command, system.MaxCmdLen())
			execs = append(execs, commands...)
		} else {
			args = append(args, b.opts.GitArgs...)
			execs = append(execs, strings.Join(args, " "))
		}
	}

	if !scriptExists {
		return nil, nil, scriptNotExistsError{params.Script}
	}

	return execs, nil, nil
}
