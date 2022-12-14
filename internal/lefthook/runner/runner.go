package runner

import (
	"bytes"
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

type Opts struct {
	Fs           afero.Fs
	Repo         *git.Repository
	Hook         *config.Hook
	GitArgs      []string
	ResultChan   chan Result
	SkipSettings log.SkipSettings
	DisableTTY   bool
	Follow       bool
}

// Runner responds for actual execution and handling the results.
type Runner struct {
	Opts

	failed   atomic.Bool
	executor Executor
}

func NewRunner(opts Opts) *Runner {
	return &Runner{
		Opts:     opts,
		executor: CommandExecutor{},
	}
}

// RunAll runs scripts and commands.
// LFS hook is executed at first if needed.
func (r *Runner) RunAll(hookName string, sourceDirs []string) {
	if err := r.runLFSHook(hookName); err != nil {
		log.Error(err)
	}

	if r.Hook.Skip != nil && r.Hook.DoSkip(r.Repo.State()) {
		logSkip(hookName, "(SKIP BY HOOK SETTING)")
		return
	}

	if !r.DisableTTY && !r.Follow {
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
	r.ResultChan <- resultFail(name, text)
	r.failed.Store(true)
}

func (r *Runner) success(name string) {
	r.ResultChan <- resultSuccess(name)
}

func (r *Runner) runLFSHook(hookName string) error {
	if !git.IsLFSHook(hookName) {
		return nil
	}

	lfsRequiredFile := filepath.Join(r.Repo.RootPath, git.LFSRequiredFile)
	lfsConfigFile := filepath.Join(r.Repo.RootPath, git.LFSConfigFile)

	requiredExists, err := afero.Exists(r.Fs, lfsRequiredFile)
	if err != nil {
		return err
	}
	configExists, err := afero.Exists(r.Fs, lfsConfigFile)
	if err != nil {
		return err
	}

	if git.IsLFSAvailable() {
		log.Debugf(
			"[git-lfs] executing hook: git lfs %s %s", hookName, strings.Join(r.GitArgs, " "),
		)
		out, err := r.executor.RawExecute(
			"git",
			append(
				[]string{"lfs", hookName},
				r.GitArgs...,
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
	files, err := afero.ReadDir(r.Fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return
	}

	interactiveScripts := make([]os.FileInfo, 0)
	var wg sync.WaitGroup

	for _, file := range files {
		script, ok := r.Hook.Scripts[file.Name()]
		if !ok {
			logSkip(file.Name(), "(SKIP BY NOT EXIST IN CONFIG)")
			continue
		}

		if r.failed.Load() && r.Hook.Piped {
			logSkip(file.Name(), "(SKIP BY BROKEN PIPE)")
			continue
		}

		if script.Interactive {
			interactiveScripts = append(interactiveScripts, file)
			continue
		}

		path := filepath.Join(dir, file.Name())

		if r.Hook.Parallel {
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
		script := r.Hook.Scripts[file.Name()]
		if r.failed.Load() {
			logSkip(file.Name(), "(SKIP INTERACTIVE BY FAILED)")
			continue
		}

		path := filepath.Join(dir, file.Name())

		r.runScript(script, path, file)
	}
}

func (r *Runner) runScript(script *config.Script, path string, file os.FileInfo) {
	if script.Skip != nil && script.DoSkip(r.Repo.State()) {
		logSkip(file.Name(), "(SKIP BY SETTINGS)")
		return
	}

	if intersect(r.Hook.ExcludeTags, script.Tags) {
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
		if err := r.Fs.Chmod(path, executableFileMode); err != nil {
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
	args = append(args, r.GitArgs[:]...)

	if script.Interactive && !r.DisableTTY && !r.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	r.run(ExecuteOptions{
		name:        file.Name(),
		root:        r.Repo.RootPath,
		args:        args,
		failText:    script.FailText,
		interactive: script.Interactive && !r.DisableTTY,
		env:         script.Env,
		follow:      r.Follow,
	})
}

func (r *Runner) runCommands() {
	commands := make([]string, 0, len(r.Hook.Commands))
	for name := range r.Hook.Commands {
		commands = append(commands, name)
	}

	sort.Strings(commands)

	interactiveCommands := make([]string, 0)
	var wg sync.WaitGroup

	for _, name := range commands {
		if r.failed.Load() && r.Hook.Piped {
			logSkip(name, "(SKIP BY BROKEN PIPE)")
			continue
		}

		if r.Hook.Commands[name].Interactive {
			interactiveCommands = append(interactiveCommands, name)
			continue
		}

		if r.Hook.Parallel {
			wg.Add(1)
			go func(name string, command *config.Command) {
				defer wg.Done()
				r.runCommand(name, command)
			}(name, r.Hook.Commands[name])
		} else {
			r.runCommand(name, r.Hook.Commands[name])
		}
	}

	wg.Wait()

	for _, name := range interactiveCommands {
		if r.failed.Load() {
			logSkip(name, "(SKIP INTERACTIVE BY FAILED)")
			continue
		}

		r.runCommand(name, r.Hook.Commands[name])
	}
}

func (r *Runner) runCommand(name string, command *config.Command) {
	if command.Skip != nil && command.DoSkip(r.Repo.State()) {
		logSkip(name, "(SKIP BY SETTINGS)")
		return
	}

	if intersect(r.Hook.ExcludeTags, command.Tags) {
		logSkip(name, "(SKIP BY TAGS)")
		return
	}

	if intersect(r.Hook.ExcludeTags, []string{name}) {
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

	if command.Interactive && !r.DisableTTY && !r.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	r.run(ExecuteOptions{
		name:        name,
		root:        filepath.Join(r.Repo.RootPath, command.Root),
		args:        args,
		failText:    command.FailText,
		interactive: command.Interactive && !r.DisableTTY,
		env:         command.Env,
		follow:      r.Follow,
	})
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

func (r *Runner) run(opts ExecuteOptions) {
	log.SetName(opts.name)
	defer log.UnsetName(opts.name)

	if (opts.follow || opts.interactive) && !r.SkipSettings.SkipExecution() {
		log.Info(log.Cyan("\n  EXECUTE > "), log.Bold(opts.name))
		err := r.executor.Execute(opts, os.Stdout)
		if err != nil {
			r.fail(opts.name, opts.failText)
		} else {
			r.success(opts.name)
		}
		return
	}

	out := bytes.NewBuffer(make([]byte, 0))
	err := r.executor.Execute(opts, out)

	var execName string
	if err != nil {
		r.fail(opts.name, opts.failText)
		execName = fmt.Sprint(log.Red("\n  EXECUTE > "), log.Bold(opts.name))
	} else {
		r.success(opts.name)
		execName = fmt.Sprint(log.Cyan("\n  EXECUTE > "), log.Bold(opts.name))
	}

	if opts.follow {
		return
	}

	if err == nil && r.SkipSettings.SkipExecution() {
		return
	}

	log.Infof("%s\n%s", execName, out)
	if err != nil {
		log.Infof("%s", err)
	}
	log.Infof("\n")
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
