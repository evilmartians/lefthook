package runner

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

func (r *Runner) prepareCommand(name string, command *config.Command) ([]string, error) {
	if command.Skip != nil && command.DoSkip(r.Repo.State()) {
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
		return nil, errors.New("invalid conig")
	}

	args, err := r.buildCommandArgs(command)
	if err != nil {
		log.Error(err)
		return nil, errors.New("error")
	}
	if len(args) == 0 {
		return nil, errors.New("no files for inspection")
	}

	return args, nil
}

func (r *Runner) buildCommandArgs(command *config.Command) ([]string, error) {
	filesCommand := r.Hook.Files
	if command.Files != "" {
		filesCommand = command.Files
	}

	filesTypeToFn := map[string]func() ([]string, error){
		config.SubStagedFiles: r.Repo.StagedFiles,
		config.PushFiles:      r.Repo.PushFiles,
		config.SubAllFiles:    r.Repo.AllFiles,
		config.SubFiles: func() ([]string, error) {
			return r.Repo.FilesByCommand(filesCommand)
		},
	}

	runString := command.Run
	for filesType, filesFn := range filesTypeToFn {
		// Checking substitutions and skipping execution if it is empty.
		//
		// Special case - `files` option: return if the result of files
		// command is empty.
		if strings.Contains(runString, filesType) ||
			filesCommand != "" && filesType == config.SubFiles {
			files, err := filesFn()
			if err != nil {
				return nil, fmt.Errorf("error replacing %s: %w", filesType, err)
			}
			if len(files) == 0 {
				return nil, nil
			}

			filesPrepared := prepareFiles(command, files)
			if len(filesPrepared) == 0 {
				return nil, nil
			}

			runString = replaceQuoted(runString, filesType, filesPrepared)
		}
	}

	runString = strings.ReplaceAll(runString, "{0}", strings.Join(r.GitArgs, " "))
	for i, gitArg := range r.GitArgs {
		runString = strings.ReplaceAll(runString, fmt.Sprintf("{%d}", i+1), gitArg)
	}

	log.Debug("[lefthook] executing: ", runString)

	return strings.Split(runString, " "), nil
}

func prepareFiles(command *config.Command, files []string) []string {
	if files == nil {
		return []string{}
	}

	log.Debug("[lefthook] files before filters:\n", files)

	files = filterGlob(files, command.Glob)
	files = filterExclude(files, command.Exclude)
	files = filterRelative(files, command.Root)

	log.Debug("[lefthook] files after filters:\n", files)

	// Escape file names to prevent unexpected bugs
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
