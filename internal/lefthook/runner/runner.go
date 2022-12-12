package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/afero"
	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
)

type status int8

const (
	executableFileMode os.FileMode = 0o751
	executableMask     os.FileMode = 0o111
)

var surroundingQuotesRegexp = regexp.MustCompile(`^'(.*)'$`)

// Runner responds for actual execution and handling the results.
type Runner struct {
	fs             afero.Fs
	repo           *git.Repository
	hook           *config.Hook
	args           []string
	failed         atomic.Bool
	resultChan     chan Result
	exec           Executor
	logSettings    log.SkipSettings
	spinnerEnabled bool
}

func NewRunner(
	fs afero.Fs,
	repo *git.Repository,
	hook *config.Hook,
	args []string,
	resultChan chan Result,
	logSettings log.SkipSettings,
	spinnerEnabled bool,
) *Runner {
	return &Runner{
		fs:             fs,
		repo:           repo,
		hook:           hook,
		args:           args,
		resultChan:     resultChan,
		exec:           CommandExecutor{},
		logSettings:    logSettings,
		spinnerEnabled: spinnerEnabled,
	}
}

// RunAll runs scripts and commands.
// LFS hook is executed at first if needed.
func (r *Runner) RunAll(hookName string, sourceDirs []string) {
	if err := r.runLFSHook(hookName); err != nil {
		log.Error(err)
	}

	if r.hook.Skip != nil && r.hook.DoSkip(r.repo.State()) {
		logSkip(hookName, "(SKIP BY HOOK SETTING)")
		return
	}

	if r.spinnerEnabled {
		log.StartSpinner()
		defer log.StopSpinner()
	}

	scriptDirs := make([]string, len(sourceDirs))
	for _, sourceDir := range sourceDirs {
		scriptDirs = append(scriptDirs, filepath.Join(
			sourceDir, hookName,
		))
	}

	for _, dir := range scriptDirs {
		r.runScripts(dir)
	}

	r.runCommands()
}

func (r *Runner) fail(name, text string) {
	r.resultChan <- resultFail(name, text)
	r.failed.Store(true)
}

func (r *Runner) success(name string) {
	r.resultChan <- resultSuccess(name)
}

func (r *Runner) runLFSHook(hookName string) error {
	if !git.IsLFSHook(hookName) {
		return nil
	}

	lfsRequiredFile := filepath.Join(r.repo.RootPath, git.LFSRequiredFile)
	lfsConfigFile := filepath.Join(r.repo.RootPath, git.LFSConfigFile)

	requiredExists, err := afero.Exists(r.repo.Fs, lfsRequiredFile)
	if err != nil {
		return err
	}
	configExists, err := afero.Exists(r.repo.Fs, lfsConfigFile)
	if err != nil {
		return err
	}

	if git.IsLFSAvailable() {
		log.Debugf(
			"[git-lfs] executing hook: git lfs %s %s", hookName, strings.Join(r.args, " "),
		)
		out, err := r.exec.RawExecute(
			"git",
			append(
				[]string{"lfs", hookName},
				r.args...,
			)...,
		)

		output := strings.Trim(out.String(), "\n")
		if output != "" {
			log.Debug("[git-lfs] output: ", output)
		}
		if err != nil {
			log.Debug("[git-lfs] error: ", err)
		}

		if err == nil && output != "" {
			log.Info(output)
		}

		if err != nil && (requiredExists || configExists) {
			log.Warn(output)
			return fmt.Errorf("git-lfs command failed: %w", err)
		}

		return nil
	}

	if requiredExists || configExists {
		log.Errorf(
			"This repository requires Git LFS, but 'git-lfs' wasn't found.\n"+
				"Install 'git-lfs' or consider reviewing the files:\n"+
				"  - %s\n"+
				"  - %s\n",
			lfsRequiredFile, lfsConfigFile,
		)
		return errors.New("git-lfs is required")
	}

	return nil
}

func (r *Runner) runScripts(dir string) {
	files, err := afero.ReadDir(r.fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return
	}

	interactiveScripts := make([]os.FileInfo, 0)
	var wg sync.WaitGroup

	for _, file := range files {
		script, ok := r.hook.Scripts[file.Name()]
		if !ok {
			logSkip(file.Name(), "(SKIP BY NOT EXIST IN CONFIG)")
			continue
		}

		if r.failed.Load() && r.hook.Piped {
			logSkip(file.Name(), "(SKIP BY BROKEN PIPE)")
			continue
		}

		if script.Interactive {
			interactiveScripts = append(interactiveScripts, file)
			continue
		}

		path := filepath.Join(dir, file.Name())

		if r.hook.Parallel {
			wg.Add(1)
			go func(script *config.Script, path string, file os.FileInfo) {
				defer wg.Done()
				r.runScript(script, path, file)
			}(script, path, file)
		} else {
			r.runScript(script, path, file)
		}
	}

	wg.Wait()

	for _, file := range interactiveScripts {
		script := r.hook.Scripts[file.Name()]
		if r.failed.Load() {
			logSkip(file.Name(), "(SKIP INTERACTIVE BY FAILED)")
			continue
		}

		path := filepath.Join(dir, file.Name())

		r.runScript(script, path, file)
	}
}

func (r *Runner) runScript(script *config.Script, path string, file os.FileInfo) {
	if script.Skip != nil && script.DoSkip(r.repo.State()) {
		logSkip(file.Name(), "(SKIP BY SETTINGS)")
		return
	}

	if intersect(r.hook.ExcludeTags, script.Tags) {
		logSkip(file.Name(), "(SKIP BY TAGS)")
		return
	}

	// Skip non-regular files (dirs, symlinks, sockets, etc.)
	if !file.Mode().IsRegular() {
		log.Debugf("[lefthook] file %s is not a regular file, skipping", file.Name())
		return
	}

	// Make sure file is executable
	if (file.Mode() & executableMask) == 0 {
		if err := r.fs.Chmod(path, executableFileMode); err != nil {
			log.Errorf("Couldn't change file mode to make file executable: %s", err)
			r.fail(file.Name(), "")
			return
		}
	}

	var args []string
	if len(script.Runner) > 0 {
		args = strings.Split(script.Runner, " ")
	}

	args = append(args, path)
	args = append(args, r.args[:]...)

	if script.Interactive && r.spinnerEnabled {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	r.run(ExecuteOptions{
		name:        file.Name(),
		root:        r.repo.RootPath,
		args:        args,
		failText:    script.FailText,
		interactive: script.Interactive,
		env:         script.Env,
	})
}

func (r *Runner) runCommands() {
	commands := make([]string, 0, len(r.hook.Commands))
	for name := range r.hook.Commands {
		commands = append(commands, name)
	}

	sort.Strings(commands)

	interactiveCommands := make([]string, 0)
	var wg sync.WaitGroup

	for _, name := range commands {
		if r.failed.Load() && r.hook.Piped {
			logSkip(name, "(SKIP BY BROKEN PIPE)")
			continue
		}

		if r.hook.Commands[name].Interactive {
			interactiveCommands = append(interactiveCommands, name)
			continue
		}

		if r.hook.Parallel {
			wg.Add(1)
			go func(name string, command *config.Command) {
				defer wg.Done()
				r.runCommand(name, command)
			}(name, r.hook.Commands[name])
		} else {
			r.runCommand(name, r.hook.Commands[name])
		}
	}

	wg.Wait()

	for _, name := range interactiveCommands {
		if r.failed.Load() {
			logSkip(name, "(SKIP INTERACTIVE BY FAILED)")
			continue
		}

		r.runCommand(name, r.hook.Commands[name])
	}
}

func (r *Runner) runCommand(name string, command *config.Command) {
	if command.Skip != nil && command.DoSkip(r.repo.State()) {
		logSkip(name, "(SKIP BY SETTINGS)")
		return
	}

	if intersect(r.hook.ExcludeTags, command.Tags) {
		logSkip(name, "(SKIP BY TAGS)")
		return
	}

	if intersect(r.hook.ExcludeTags, []string{name}) {
		logSkip(name, "(SKIP BY NAME)")
		return
	}

	if err := command.Validate(); err != nil {
		r.fail(name, "")
		return
	}

	args, err := r.buildCommandArgs(command)
	if err != nil {
		log.Error(err)
		logSkip(name, "(SKIP. ERROR)")
		return
	}
	if len(args) == 0 {
		logSkip(name, "(SKIP. NO FILES FOR INSPECTION)")
		return
	}

	if command.Interactive && r.spinnerEnabled {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	r.run(ExecuteOptions{
		name:        name,
		root:        filepath.Join(r.repo.RootPath, command.Root),
		args:        args,
		failText:    command.FailText,
		interactive: command.Interactive,
		env:         command.Env,
	})
}

func (r *Runner) buildCommandArgs(command *config.Command) ([]string, error) {
	filesCommand := r.hook.Files
	if command.Files != "" {
		filesCommand = command.Files
	}

	filesTypeToFn := map[string]func() ([]string, error){
		config.SubStagedFiles: r.repo.StagedFiles,
		config.PushFiles:      r.repo.PushFiles,
		config.SubAllFiles:    r.repo.AllFiles,
		config.SubFiles: func() ([]string, error) {
			return r.repo.FilesByCommand(filesCommand)
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
				return nil, fmt.Errorf("error replacing %s: %s", filesType, err)
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

	runString = strings.ReplaceAll(runString, "{0}", strings.Join(r.args, " "))
	for i, gitArg := range r.args {
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

func (r *Runner) run(opts ExecuteOptions) {
	log.SetName(opts.name)
	defer log.UnsetName(opts.name)

	out, err := r.exec.Execute(opts)

	var execName string
	if err != nil {
		r.fail(opts.name, opts.failText)
		execName = fmt.Sprint(log.Red("\n  EXECUTE >"), log.Bold(opts.name))
	} else {
		r.success(opts.name)
		execName = fmt.Sprint(log.Cyan("\n  EXECUTE >"), log.Bold(opts.name))
	}

	if out != nil {
		if err == nil && r.logSettings.SkipExecution() {
			return
		}

		log.Infof("%s\n%s\n", execName, out)
	} else if err != nil {
		log.Infof("%s\n%s\n", execName, err)
	}
}

// Returns whether two arrays have at least one similar element.
func intersect(a, b []string) bool {
	intersections := make(map[string]struct{}, len(a))

	for _, v := range a {
		intersections[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := intersections[v]; ok {
			return true
		}
	}

	return false
}

func logSkip(name, reason string) {
	log.Info(fmt.Sprintf("%s: %s", log.Bold(name), log.Yellow(reason)))
}
