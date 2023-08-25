package runner

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

// An object that described the single command's run option.
type run struct {
	commands [][]string
	files    []string
}

func (r *Runner) prepareCommand(name string, command *config.Command) (*run, error) {
	if command.DoSkip(r.Repo.State()) {
		return nil, errors.New("settings")
	}

	if intersect(r.Hook.ExcludeTags, command.Tags) {
		return nil, errors.New("tags")
	}

	if intersect(r.Hook.ExcludeTags, []string{name}) {
		return nil, errors.New("name")
	}

	if err := command.Validate(); err != nil {
		r.fail(name, "")
		return nil, errors.New("invalid config")
	}

	args, err, skipReason := r.buildRun(command)
	if err != nil {
		log.Error(err)
		return nil, errors.New("error")
	}
	if skipReason != nil {
		return nil, skipReason
	}

	return args, nil
}

func (r *Runner) buildRun(command *config.Command) (*run, error, error) {
	filesCmd := r.Hook.Files
	if len(command.Files) > 0 {
		filesCmd = command.Files
	}

	var stagedFiles func() ([]string, error)
	switch {
	case len(r.Files) > 0:
		stagedFiles = func() ([]string, error) { return r.Files, nil }
	case r.AllFiles:
		stagedFiles = r.Repo.AllFiles
	default:
		stagedFiles = r.Repo.StagedFiles
	}

	filesTypeToFn := map[string]func() ([]string, error){
		config.SubStagedFiles: stagedFiles,
		config.PushFiles:      r.Repo.PushFiles,
		config.SubAllFiles:    r.Repo.AllFiles,
		config.SubFiles: func() ([]string, error) {
			filesCmd = r.replacePositionalArguments(filesCmd)
			return r.Repo.FilesByCommand(filesCmd)
		},
	}

	filesFiltered := make([]string, 0)
	runString := command.Run
	for filesType, filesFn := range filesTypeToFn {
		// Checking substitutions and skipping execution if it is empty.
		//
		// Special case - `files` option: return if the result of files
		// command is empty.
		if strings.Contains(runString, filesType) || len(filesCmd) > 0 && filesType == config.SubFiles {
			files, err := filesFn()
			if err != nil {
				return nil, fmt.Errorf("error replacing %s: %w", filesType, err), nil
			}
			if len(files) == 0 {
				return nil, nil, errors.New("no files for inspection")
			}

			filtered := filterFiles(command, files)
			filesFiltered = append(filesFiltered, filtered...)

			prepared := escapeFiles(filtered)
			if len(prepared) == 0 {
				return nil, nil, errors.New("no files for inspection")
			}

			runString = replaceQuoted(runString, filesType, prepared)
		}
	}

	if len(filesFiltered) == 0 && config.HookUsesStagedFiles(r.HookName) {
		files, err := r.Repo.StagedFiles()
		if err == nil {
			if len(filterFiles(command, files)) == 0 {
				return nil, nil, errors.New("no matching staged files")
			}
		}
	}

	if len(filesFiltered) == 0 && config.HookUsesPushFiles(r.HookName) {
		files, err := r.Repo.PushFiles()
		if err == nil {
			if len(filterFiles(command, files)) == 0 {
				return nil, nil, errors.New("no matching push files")
			}
		}
	}

	runString = r.replacePositionalArguments(runString)

	log.Debug("[lefthook] executing: ", runString)

	return &run{
		commands: [][]string{strings.Split(runString, " ")},
		files:    filesFiltered,
	}, nil, nil
}

func (r *Runner) replacePositionalArguments(runString string) string {
	runString = strings.ReplaceAll(runString, "{0}", strings.Join(r.GitArgs, " "))
	for i, gitArg := range r.GitArgs {
		runString = strings.ReplaceAll(runString, fmt.Sprintf("{%d}", i+1), gitArg)
	}
	return runString
}

func filterFiles(command *config.Command, files []string) []string {
	if files == nil {
		return []string{}
	}

	log.Debug("[lefthook] files before filters:\n", files)

	files = filterGlob(files, command.Glob)
	files = filterExclude(files, command.Exclude)
	files = filterRelative(files, command.Root)

	log.Debug("[lefthook] files after filters:\n", files)

	return files
}

// Escape file names to prevent unexpected bugs.
func escapeFiles(files []string) []string {
	var filesEsc []string
	for _, fileName := range files {
		if len(fileName) > 0 {
			filesEsc = append(filesEsc, shellescape.Quote(fileName))
		}
	}

	log.Debug("[lefthook] files after escaping:\n", filesEsc)

	return filesEsc
}

func replaceQuoted(source, substitution string, files []string) string {
	for _, elem := range [][]string{
		{"\"", "\"" + substitution + "\""},
		{"'", "'" + substitution + "'"},
		{"", substitution},
	} {
		quote := elem[0]
		sub := elem[1]
		if !strings.Contains(source, sub) {
			continue
		}

		quotedFiles := files
		if len(quote) != 0 {
			quotedFiles = make([]string, 0, len(files))
			for _, fileName := range files {
				quotedFiles = append(quotedFiles,
					quote+surroundingQuotesRegexp.ReplaceAllString(fileName, "$1")+quote)
			}
		}

		source = strings.ReplaceAll(
			source, sub, strings.Join(quotedFiles, " "),
		)
	}

	return source
}
