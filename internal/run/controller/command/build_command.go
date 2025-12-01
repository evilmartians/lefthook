package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
)

var surroundingQuotesRegexp = regexp.MustCompile(`^'(.*)'$`)

// fileTemplate contains for template replacements in a command string.
type filesTemplate struct {
	files []string
	cnt   int
}

func (b *Builder) buildCommand(params *JobParams) ([]string, []string, error) {
	if err := params.validateCommand(); err != nil {
		return nil, nil, err
	}

	filesCmd := params.FilesCmd
	if len(filesCmd) > 0 {
		filesCmd = replacePositionalArguments(filesCmd, b.opts.GitArgs)
	}

	// Predefined templates
	predefined := make(map[string]string)
	for i, arg := range b.opts.GitArgs {
		predefined[strconv.Itoa(i)] = arg
	}
	for key, replacement := range b.opts.Templates {
		predefined["{"+key+"}"] = replacement
	}
	predefined["{lefthook_job_name}"] = shellescape.Quote(params.Name)

	var patterns patterns
	if len(b.opts.ForceFiles) > 0 {
		patterns = newMockedPatterns(b.opts.ForceFiles, predefined)
	} else {
		patterns = newPatterns(b.git, params.Root, filesCmd, predefined)
	}

	// var stagedFiles func() ([]string, error)
	// var stagedFilesWithDeleted func() ([]string, error)
	// var pushFiles func() ([]string, error)
	// var allFiles func() ([]string, error)
	// var cmdFiles func() ([]string, error)
	//
	// if len(b.opts.ForceFiles) > 0 {
	// 	stagedFiles = func() ([]string, error) { return b.opts.ForceFiles, nil }
	// 	stagedFilesWithDeleted = stagedFiles
	// 	pushFiles = stagedFiles
	// 	allFiles = stagedFiles
	// 	cmdFiles = stagedFiles
	// } else {
	// 	stagedFiles = b.git.StagedFiles
	// 	stagedFilesWithDeleted = b.git.StagedFilesWithDeleted
	// 	pushFiles = b.git.PushFiles
	// 	allFiles = b.git.AllFiles
	// 	cmdFiles = func() ([]string, error) {
	// 		var cmd []string
	// 		if runtime.GOOS == "windows" {
	// 			cmd = strings.Split(filesCmd, " ")
	// 		} else {
	// 			cmd = []string{"sh", "-c", filesCmd}
	// 		}
	// 		return b.git.FindExistingFiles(cmd, params.Root)
	// 	}
	// }

	// filesFns := map[string]func() ([]string, error){
	// 	config.SubStagedFiles: stagedFiles,
	// 	config.SubPushFiles:   pushFiles,
	// 	config.SubAllFiles:    allFiles,
	// 	config.SubFiles:       cmdFiles,
	// }
	//
	// filesTemplates := make(map[string]*filesTemplate)

	filter := filter.New(b.git.Fs, filter.Params{
		Glob:         params.Glob,
		ExcludeFiles: params.ExcludeFiles,
		Root:         params.Root,
		FileTypes:    params.FileTypes,
		GlobMatcher:  b.opts.GlobMatcher,
	})
	// for filesType, fn := range filesFns {
	// 	cnt := strings.Count(params.Run, filesType)
	// 	if cnt == 0 {
	// 		continue
	// 	}
	//
	// 	templ := &filesTemplate{cnt: cnt}
	// 	filesTemplates[filesType] = templ
	//
	// 	files, err := fn()
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("error replacing %s: %w", filesType, err)
	// 	}
	//
	// 	files = filter.Apply(files)
	// 	if !b.opts.Force && len(files) == 0 {
	// 		return nil, nil, SkipError{"no files for inspection"}
	// 	}
	//
	// 	templ.files = files
	// }

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case when `files` option specified but not referenced in `run`: return if the result is empty.
	// if !b.opts.Force && len(filesCmd) > 0 && filesTemplates[config.SubFiles] == nil {
	// 	files, err := filesFns[config.SubFiles]()
	// 	if err != nil {
	// 		return nil, nil, fmt.Errorf("error calling replace command for %s: %w", config.SubFiles, err)
	// 	}
	//
	// 	files = filter.Apply(files)
	//
	// 	if len(files) == 0 {
	// 		return nil, nil, SkipError{"no files for inspection"}
	// 	}
	// }

	// runString := params.Run
	// runString = replacePositionalArguments(runString, b.opts.GitArgs)
	//
	// for keyword, replacement := range b.opts.Templates {
	// 	runString = strings.ReplaceAll(runString, "{"+keyword+"}", replacement)
	// }
	//
	// runString = strings.ReplaceAll(runString, "{lefthook_job_name}", shellescape.Quote(params.Name))

	// Cache entries
	err := patterns.Discover(params.Run, filter)
	if err != nil {
		return nil, nil, err
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case when `files` option specified but not referenced in `run`: return if the result is empty.
	if !b.opts.Force && len(filesCmd) > 0 && patterns.Empty(config.SubFiles) {
		files, err := patterns.Files(config.SubFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			return nil, nil, SkipError{"no files for inspection"}
		}
	}

	// maxlen := system.MaxCmdLen()
	// commands, files := replaceInChunks(runString, filesTemplates, maxlen)

	commands, files := patterns.ReplaceAndSplit(params.Run)

	if b.opts.Force || len(files) != 0 {
		return commands, files, nil
	}

	// if config.HookUsesStagedFiles(b.opts.HookName) {
	// 	ok, err := canSkipJob(filter, filesTemplates[config.SubStagedFiles], stagedFilesWithDeleted)
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// 	if ok {
	// 		return nil, nil, SkipError{"no matching staged files"}
	// 	}
	// }
	//
	// if config.HookUsesPushFiles(b.opts.HookName) {
	// 	ok, err := canSkipJob(filter, filesTemplates[config.SubPushFiles], pushFiles)
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// 	if ok {
	// 		return nil, nil, SkipError{"no matching push files"}
	// 	}
	// }

	// Skip if no files were staged (including deleted)
	if config.HookUsesStagedFiles(b.opts.HookName) {
		files, err := patterns.Files(config.SubStagedFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			files, err = b.git.StagedFilesWithDeleted()
			if err != nil {
				return nil, nil, err
			}

			if len(filter.Apply(files)) == 0 {
				return nil, nil, SkipError{"no matching staged files"}
			}
		}
	}

	// Skip if no files were to be pushed
	if config.HookUsesPushFiles(b.opts.HookName) {
		files, err := patterns.Files(config.SubPushFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			return nil, nil, SkipError{"no matching push files"}
		}
	}

	return commands, files, nil
}

// func canSkipJob(filter *filter.Filter, template *filesTemplate, filesFn func() ([]string, error)) (bool, error) {
// 	if template != nil {
// 		return len(template.files) == 0, nil
// 	}
//
// 	files, err := filesFn()
// 	if err != nil {
// 		return false, fmt.Errorf("error getting files: %w", err)
// 	}
// 	if len(filter.Apply(files)) == 0 {
// 		return true, nil
// 	}
//
// 	return false, nil
// }

func replacePositionalArguments(str string, args []string) string {
	str = strings.ReplaceAll(str, "{0}", strings.Join(args, " "))
	for i, arg := range args {
		str = strings.ReplaceAll(str, fmt.Sprintf("{%d}", i+1), arg)
	}
	return str
}

func replaceInChunks(str string, templates map[string]*filesTemplate, maxlen int) ([]string, []string) {
	if len(templates) == 0 {
		return []string{str}, nil
	}

	var cnt int

	allFiles := make([]string, 0)
	for name, template := range templates {
		if template.cnt == 0 {
			continue
		}

		cnt += template.cnt
		maxlen += template.cnt * len(name)
		allFiles = append(allFiles, template.files...)
		template.files = escapeFiles(template.files)
	}

	maxlen -= len(str)

	if cnt > 0 {
		maxlen /= cnt
	}

	var exhausted int
	commands := make([]string, 0)
	for {
		command := str
		for name, template := range templates {
			added, rest := getNChars(template.files, maxlen)
			if len(rest) == 0 {
				exhausted += 1
			} else {
				template.files = rest
			}
			command = replaceQuoted(command, name, added)
		}

		log.Debug("[lefthook] job: ", command)
		commands = append(commands, command)
		if exhausted >= len(templates) {
			break
		}
	}

	return commands, allFiles
}

// // Escape file names to prevent unexpected bugs.
// func escapeFiles(files []string) []string {
// 	var filesEsc []string
// 	for _, fileName := range files {
// 		if len(fileName) > 0 {
// 			filesEsc = append(filesEsc, shellescape.Quote(fileName))
// 		}
// 	}
//
// 	log.Builder(log.DebugLevel, "[lefthook] ").
// 		Add("files after escaping: ", filesEsc).
// 		Log()
//
// 	return filesEsc
// }
//
// func getNChars(s []string, n int) ([]string, []string) {
// 	if len(s) == 0 {
// 		return nil, nil
// 	}
//
// 	var cnt int
// 	for i, str := range s {
// 		cnt += len(str)
// 		if i > 0 {
// 			cnt += 1 // a space
// 		}
// 		if cnt > n {
// 			if i == 0 {
// 				i = 1
// 			}
// 			return s[:i], s[i:]
// 		}
// 	}
//
// 	return s, nil
// }
//
// func replaceQuoted(source, substitution string, files []string) string {
// 	for _, elem := range [][]string{
// 		{"\"", "\"" + substitution + "\""},
// 		{"'", "'" + substitution + "'"},
// 		{"", substitution},
// 	} {
// 		quote := elem[0]
// 		sub := elem[1]
// 		if !strings.Contains(source, sub) {
// 			continue
// 		}
//
// 		quotedFiles := files
// 		if len(quote) != 0 {
// 			quotedFiles = make([]string, 0, len(files))
// 			for _, fileName := range files {
// 				quotedFiles = append(quotedFiles,
// 					quote+surroundingQuotesRegexp.ReplaceAllString(fileName, "$1")+quote)
// 			}
// 		}
//
// 		source = strings.ReplaceAll(
// 			source, sub, strings.Join(quotedFiles, " "),
// 		)
// 	}
//
// 	return source
// }
