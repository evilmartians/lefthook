package runner

import (
	"fmt"
	"runtime"
	"strings"

	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/filters"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
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

func (r *Runner) prepareCommand(name string, command *config.Command) (*run, error) {
	if command.DoSkip(r.Repo.State) {
		return nil, &skipError{"settings"}
	}

	if intersect(r.Hook.ExcludeTags, command.Tags) {
		return nil, &skipError{"tags"}
	}

	if intersect(r.Hook.ExcludeTags, []string{name}) {
		return nil, &skipError{"name"}
	}

	if err := command.Validate(); err != nil {
		return nil, err
	}

	args, err := r.buildRun(command)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func (r *Runner) buildRun(command *config.Command) (*run, error) {
	filesCmd := r.Hook.Files
	if len(command.Files) > 0 {
		filesCmd = command.Files
	}
	if len(filesCmd) > 0 {
		filesCmd = replacePositionalArguments(filesCmd, r.GitArgs)
	}

	var stagedFiles func() ([]string, error)
	var pushFiles func() ([]string, error)
	var allFiles func() ([]string, error)
	var cmdFiles func() ([]string, error)

	if len(r.Files) > 0 {
		stagedFiles = func() ([]string, error) { return r.Files, nil }
		pushFiles = stagedFiles
		allFiles = stagedFiles
		cmdFiles = stagedFiles
	} else {
		stagedFiles = r.Repo.StagedFiles
		pushFiles = r.Repo.PushFiles
		allFiles = r.Repo.AllFiles
		cmdFiles = func() ([]string, error) {
			var cmd []string
			if runtime.GOOS == "windows" {
				cmd = strings.Split(filesCmd, " ")
			} else {
				cmd = []string{"sh", "-c", filesCmd}
			}
			return r.Repo.FilesByCommand(cmd)
		}
	}

	filesFns := map[string]func() ([]string, error){
		config.SubStagedFiles: stagedFiles,
		config.SubPushFiles:   pushFiles,
		config.SubAllFiles:    allFiles,
		config.SubFiles:       cmdFiles,
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
			return nil, fmt.Errorf("error replacing %s: %w", filesType, err)
		}

		files = filters.Apply(r.Repo.Fs, command, files)
		if !r.Force && len(files) == 0 {
			return nil, &skipError{"no files for inspection"}
		}

		templ.files = files
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case for `files` option: return if the result of files command is empty.
	if !r.Force && len(filesCmd) > 0 && templates[config.SubFiles] == nil {
		files, err := filesFns[config.SubFiles]()
		if err != nil {
			return nil, fmt.Errorf("error calling replace command for %s: %w", config.SubFiles, err)
		}

		files = filters.Apply(r.Repo.Fs, command, files)

		if len(files) == 0 {
			return nil, &skipError{"no files for inspection"}
		}
	}

	runString := command.Run
	runString = replacePositionalArguments(runString, r.GitArgs)

	maxlen := system.MaxCmdLen()
	result := replaceInChunks(runString, templates, maxlen)

	if r.Force || len(result.files) != 0 {
		return result, nil
	}

	if config.HookUsesStagedFiles(r.HookName) {
		ok, err := r.canSkipCommand(command, templates[config.SubStagedFiles], r.Repo.StagedFiles)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, &skipError{"no matching staged files"}
		}
	}

	if config.HookUsesPushFiles(r.HookName) {
		ok, err := r.canSkipCommand(command, templates[config.SubPushFiles], r.Repo.PushFiles)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, &skipError{"no matching push files"}
		}
	}

	return result, nil
}

func (r *Runner) canSkipCommand(command *config.Command, template *template, filesFn func() ([]string, error)) (bool, error) {
	if template != nil {
		return len(template.files) == 0, nil
	}

	files, err := filesFn()
	if err != nil {
		return false, fmt.Errorf("error getting files: %w", err)
	}
	if len(filters.Apply(r.Repo.Fs, command, files)) == 0 {
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
