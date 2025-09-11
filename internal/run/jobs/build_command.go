package jobs

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/filters"
	"github.com/evilmartians/lefthook/internal/system"
)

var surroundingQuotesRegexp = regexp.MustCompile(`^'(.*)'$`)

// fileTemplate contains for template replacements in a command string.
type filesTemplate struct {
	files []string
	cnt   int
}

func buildCommand(params *Params, settings *Settings) (*Job, error) {
	if err := params.validateCommand(); err != nil {
		return nil, err
	}

	filesCmd := settings.Hook.Files
	if len(params.Files) > 0 {
		filesCmd = params.Files
	}
	if len(filesCmd) > 0 {
		filesCmd = replacePositionalArguments(filesCmd, settings.GitArgs)
	}

	var stagedFiles func() ([]string, error)
	var stagedFilesWithDeleted func() ([]string, error)
	var pushFiles func() ([]string, error)
	var allFiles func() ([]string, error)
	var cmdFiles func() ([]string, error)

	if len(settings.ForceFiles) > 0 {
		stagedFiles = func() ([]string, error) { return settings.ForceFiles, nil }
		stagedFilesWithDeleted = stagedFiles
		pushFiles = stagedFiles
		allFiles = stagedFiles
		cmdFiles = stagedFiles
	} else {
		stagedFiles = settings.Repo.StagedFiles
		stagedFilesWithDeleted = settings.Repo.StagedFilesWithDeleted
		pushFiles = settings.Repo.PushFiles
		allFiles = settings.Repo.AllFiles
		cmdFiles = func() ([]string, error) {
			var cmd []string
			if runtime.GOOS == "windows" {
				cmd = strings.Split(filesCmd, " ")
			} else {
				cmd = []string{"sh", "-c", filesCmd}
			}
			return settings.Repo.FindExistingFiles(cmd, params.Root)
		}
	}

	filesFns := map[string]func() ([]string, error){
		config.SubStagedFiles: stagedFiles,
		config.SubPushFiles:   pushFiles,
		config.SubAllFiles:    allFiles,
		config.SubFiles:       cmdFiles,
	}

	filesTemplates := make(map[string]*filesTemplate)

	filterParams := filters.Params{
		Glob:      params.Glob,
		Exclude:   params.Exclude,
		Root:      params.Root,
		FileTypes: params.FileTypes,
	}
	for filesType, fn := range filesFns {
		cnt := strings.Count(params.Run, filesType)
		if cnt == 0 {
			continue
		}

		templ := &filesTemplate{cnt: cnt}
		filesTemplates[filesType] = templ

		files, err := fn()
		if err != nil {
			return nil, fmt.Errorf("error replacing %s: %w", filesType, err)
		}

		files = filters.Apply(settings.Repo.Fs, files, filterParams)
		if !settings.Force && len(files) == 0 {
			return nil, SkipError{"no files for inspection"}
		}

		templ.files = files
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case for `files` option: return if the result of files command is empty.
	if !settings.Force && len(filesCmd) > 0 && filesTemplates[config.SubFiles] == nil {
		files, err := filesFns[config.SubFiles]()
		if err != nil {
			return nil, fmt.Errorf("error calling replace command for %s: %w", config.SubFiles, err)
		}

		files = filters.Apply(settings.Repo.Fs, files, filterParams)

		if len(files) == 0 {
			return nil, SkipError{"no files for inspection"}
		}
	}

	runString := params.Run
	runString = replacePositionalArguments(runString, settings.GitArgs)

	for keyword, replacement := range settings.Templates {
		runString = strings.ReplaceAll(runString, "{"+keyword+"}", replacement)
	}

	runString = strings.ReplaceAll(runString, "{lefthook_job_name}", shellescape.Quote(params.Name))

	maxlen := system.MaxCmdLen()
	result := replaceInChunks(runString, filesTemplates, maxlen)

	if settings.Force || len(result.Files) != 0 {
		return result, nil
	}

	if config.HookUsesStagedFiles(settings.HookName) {
		ok, err := canSkipJob(settings, filterParams, filesTemplates[config.SubStagedFiles], stagedFilesWithDeleted)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, SkipError{"no matching staged files"}
		}
	}

	if config.HookUsesPushFiles(settings.HookName) {
		ok, err := canSkipJob(settings, filterParams, filesTemplates[config.SubPushFiles], pushFiles)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, SkipError{"no matching push files"}
		}
	}

	return result, nil
}

func canSkipJob(settings *Settings, filterParams filters.Params, template *filesTemplate, filesFn func() ([]string, error)) (bool, error) {
	if template != nil {
		return len(template.files) == 0, nil
	}

	files, err := filesFn()
	if err != nil {
		return false, fmt.Errorf("error getting files: %w", err)
	}
	if len(filters.Apply(settings.Repo.Fs, files, filterParams)) == 0 {
		return true, nil
	}

	return false, nil
}

func replacePositionalArguments(str string, args []string) string {
	str = strings.ReplaceAll(str, "{0}", strings.Join(args, " "))
	for i, arg := range args {
		str = strings.ReplaceAll(str, fmt.Sprintf("{%d}", i+1), arg)
	}
	return str
}

// Escape file names to prevent unexpected bugs.
func escapeFiles(files []string) []string {
	var filesEsc []string
	for _, fileName := range files {
		if len(fileName) > 0 {
			filesEsc = append(filesEsc, shellescape.Quote(fileName))
		}
	}

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("files after escaping: ", filesEsc).
		Log()

	return filesEsc
}

func replaceInChunks(str string, templates map[string]*filesTemplate, maxlen int) *Job {
	if len(templates) == 0 {
		return &Job{
			Execs: []string{str},
		}
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

	return &Job{
		Execs: commands,
		Files: allFiles,
	}
}

func getNChars(s []string, n int) ([]string, []string) {
	if len(s) == 0 {
		return nil, nil
	}

	var cnt int
	for i, str := range s {
		cnt += len(str)
		if i > 0 {
			cnt += 1 // a space
		}
		if cnt > n {
			if i == 0 {
				i = 1
			}
			return s[:i], s[i:]
		}
	}

	return s, nil
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
