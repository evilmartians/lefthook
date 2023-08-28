package runner

import (
	"errors"
	"fmt"
	"runtime"
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

type template struct {
	files []string
	cnt   int
}

const (
	// https://serverfault.com/questions/69430/what-is-the-maximum-length-of-a-command-line-in-mac-os-x
	// https://support.microsoft.com/en-us/help/830473/command-prompt-cmd-exe-command-line-string-limitation
	// https://unix.stackexchange.com/a/120652
	maxCommandLengthDarwin  = 260000 // 262144
	maxCommandLengthWindows = 8000   // 8191
	maxCommandLengthLinux   = 130000 // 131072
)

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
		r.fail(name, err)
		return nil, err
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

	filesFns := map[string]func() ([]string, error){
		config.SubStagedFiles: stagedFiles,
		config.PushFiles:      r.Repo.PushFiles,
		config.SubAllFiles:    r.Repo.AllFiles,
		config.SubFiles: func() ([]string, error) {
			filesCmd = r.replacePositionalArguments(filesCmd)
			return r.Repo.FilesByCommand(filesCmd)
		},
	}

	templates := make(map[string]*template)

	for filesType, fn := range filesFns {
		cnt := strings.Count(command.Run, filesType)
		if cnt == 0 {
			continue
		}

		templ := &template{cnt: cnt}
		templates[filesType] = templ

		files, err := fn()
		if err != nil {
			return nil, fmt.Errorf("error replacing %s: %w", filesType, err), nil
		}

		files = filterFiles(command, files)
		files = escapeFiles(files)
		if len(files) == 0 {
			return nil, nil, errors.New("no files for inspection")
		}

		templ.files = files
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case - `files` option: return if the result of files
	// command is empty.
	if len(filesCmd) > 0 && templates[config.SubFiles] == nil {
		files, err := filesFns[config.SubFiles]()
		if err != nil {
			return nil, fmt.Errorf("error calling replace command for %s: %w", config.SubFiles, err), nil
		}

		files = filterFiles(command, files)
		files = escapeFiles(files)

		if len(files) == 0 {
			return nil, nil, errors.New("no files for inspection")
		}
	}

	runString := command.Run
	runString = r.replacePositionalArguments(runString)
	log.Debugf("[lefthook] found templates: %+v", templates)
	result := replaceTemplates(runString, templates)

	if len(result.files) == 0 && config.HookUsesStagedFiles(r.HookName) {
		if templates[config.SubStagedFiles] != nil && len(templates[config.SubStagedFiles].files) == 0 {
			return nil, nil, errors.New("no matching staged files")
		}

		files, err := r.Repo.StagedFiles()
		if err == nil {
			if len(filterFiles(command, files)) == 0 {
				return nil, nil, errors.New("no matching staged files")
			}
		}
	}

	if len(result.files) == 0 && config.HookUsesPushFiles(r.HookName) {
		if templates[config.PushFiles] != nil && len(templates[config.PushFiles].files) == 0 {
			return nil, nil, errors.New("no matching push files")
		}

		files, err := r.Repo.PushFiles()
		if err == nil {
			if len(filterFiles(command, files)) == 0 {
				return nil, nil, errors.New("no matching push files")
			}
		}
	}

	log.Debugf("[lefthook] executing: %+v", result)

	return result, nil, nil
}

func (r *Runner) replacePositionalArguments(str string) string {
	str = strings.ReplaceAll(str, "{0}", strings.Join(r.GitArgs, " "))
	for i, gitArg := range r.GitArgs {
		str = strings.ReplaceAll(str, fmt.Sprintf("{%d}", i+1), gitArg)
	}
	return str
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

func replaceTemplates(str string, templates map[string]*template) *run {
	if len(templates) == 0 {
		return &run{
			commands: [][]string{strings.Split(str, " ")},
		}
	}

	var cnt int

	allFiles := make([]string, 0)
	for _, template := range templates {
		if template.cnt == 0 {
			continue
		}

		cnt += template.cnt
		allFiles = append(allFiles, template.files...)
	}

	var maxStringLength int
	switch runtime.GOOS {
	case "windows":
		maxStringLength = maxCommandLengthWindows
	case "darwin":
		maxStringLength = maxCommandLengthDarwin
	default:
		maxStringLength = maxCommandLengthLinux
	}

	if cnt > 0 {
		maxStringLength /= cnt
	}

	commands := make([][]string, 0)
out:
	for {
		command := str
		for name, template := range templates {
			added, rest := getNChars(template.files, maxStringLength-len(name))
			if len(added) == 0 {
				break out
			}
			command = replaceQuoted(command, name, added)
			log.Debug("[lefthook] chunk command: ", command)
			template.files = rest
			if len(rest) == 0 {
				break
			}
		}

		commands = append(commands, strings.Split(command, " "))
	}

	return &run{
		commands: commands,
		files:    allFiles,
	}
}

func getNChars(s []string, n int) ([]string, []string) {
	if len(s) == 0 {
		return nil, nil
	}

	var cnt int
	for i, str := range s {
		cnt += len(str)
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
