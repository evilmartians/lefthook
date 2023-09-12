package run

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/run/filter"
	"github.com/evilmartians/lefthook/internal/log"
)

// An object that describes the single command's run option.
type run struct {
	commands []string
	files    []string
}

// Stats for template replacements in a command string.
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
	if len(filesCmd) > 0 {
		filesCmd = replacePositionalArguments(filesCmd, r.GitArgs)
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

		files = filter.Apply(command, files)
		if len(files) == 0 {
			return nil, nil, errors.New("no files for inspection")
		}

		templ.files = files
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case for `files` option: return if the result of files command is empty.
	if len(filesCmd) > 0 && templates[config.SubFiles] == nil {
		files, err := filesFns[config.SubFiles]()
		if err != nil {
			return nil, fmt.Errorf("error calling replace command for %s: %w", config.SubFiles, err), nil
		}

		files = filter.Apply(command, files)

		if len(files) == 0 {
			return nil, nil, errors.New("no files for inspection")
		}
	}

	runString := command.Run
	runString = replacePositionalArguments(runString, r.GitArgs)

	var maxlen int
	switch runtime.GOOS {
	case "windows":
		maxlen = maxCommandLengthWindows
	case "darwin":
		maxlen = maxCommandLengthDarwin
	default:
		maxlen = maxCommandLengthLinux
	}
	result := replaceInChunks(runString, templates, maxlen)

	if len(result.files) == 0 && config.HookUsesStagedFiles(r.HookName) {
		if templates[config.SubStagedFiles] != nil && len(templates[config.SubStagedFiles].files) == 0 {
			return nil, nil, errors.New("no matching staged files")
		}

		files, err := r.Repo.StagedFiles()
		if err == nil {
			if len(filter.Apply(command, files)) == 0 {
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
			if len(filter.Apply(command, files)) == 0 {
				return nil, nil, errors.New("no matching push files")
			}
		}
	}

	return result, nil, nil
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

	log.Debug("[lefthook] files after escaping:\n", filesEsc)

	return filesEsc
}

func replaceInChunks(str string, templates map[string]*template, maxlen int) *run {
	if len(templates) == 0 {
		return &run{
			commands: []string{str},
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

		log.Debug("[lefthook] executing: ", command)
		commands = append(commands, command)
		if exhausted >= len(templates) {
			break
		}
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
