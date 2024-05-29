package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/exec"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/filters"
	"github.com/evilmartians/lefthook/internal/log"
)

const (
	executableFileMode os.FileMode = 0o751
	executableMask     os.FileMode = 0o111
	execLogPadding                 = 2
)

var surroundingQuotesRegexp = regexp.MustCompile(`^'(.*)'$`)

type Options struct {
	Repo            *git.Repository
	Hook            *config.Hook
	HookName        string
	GitArgs         []string
	LogSettings     log.Settings
	DisableTTY      bool
	Force           bool
	Files           []string
	RunOnlyCommands []string
}

// Runner responds for actual execution and handling the results.
type Runner struct {
	Options

	stdin                io.Reader
	partiallyStagedFiles []string
	failed               atomic.Bool
	executor             exec.Executor
}

func New(opts Options) *Runner {
	return &Runner{
		Options: opts,

		// Some hooks use STDIN for parsing data from Git. To allow multiple commands
		// and scripts access the same Git data STDIN is cached via cachedReader.
		stdin:    NewCachedReader(os.Stdin),
		executor: exec.CommandExecutor{},
	}
}

// skipError implements error interface but indicates that the execution needs to be skipped.
type skipError struct {
	reason string
}

func (r *skipError) Error() string {
	return r.reason
}

type executable interface {
	*config.Command | *config.Script
	ExecutionPriority() int
}

// RunAll runs scripts and commands.
// LFS hook is executed at first if needed.
func (r *Runner) RunAll(ctx context.Context, sourceDirs []string) []Result {
	results := make([]Result, 0, len(r.Hook.Commands)+len(r.Hook.Scripts))

	if err := r.runLFSHook(ctx); err != nil {
		log.Error(err)
	}

	if r.Hook.DoSkip(r.Repo.State()) {
		r.logSkip(r.HookName, "hook setting")
		return results
	}

	if !r.DisableTTY && !r.Hook.Follow {
		log.StartSpinner()
		defer log.StopSpinner()
	}

	scriptDirs := make([]string, 0, len(sourceDirs))
	for _, sourceDir := range sourceDirs {
		scriptDirs = append(scriptDirs, filepath.Join(
			sourceDir, r.HookName,
		))
	}

	r.preHook()

	for _, dir := range scriptDirs {
		results = append(results, r.runScripts(ctx, dir)...)
	}

	results = append(results, r.runCommands(ctx)...)

	r.postHook()

	return results
}

func (r *Runner) runLFSHook(ctx context.Context) error {
	if !git.IsLFSHook(r.HookName) {
		return nil
	}

	lfsRequiredFile := filepath.Join(r.Repo.RootPath, git.LFSRequiredFile)
	lfsConfigFile := filepath.Join(r.Repo.RootPath, git.LFSConfigFile)

	requiredExists, err := afero.Exists(r.Repo.Fs, lfsRequiredFile)
	if err != nil {
		return err
	}
	configExists, err := afero.Exists(r.Repo.Fs, lfsConfigFile)
	if err != nil {
		return err
	}

	if git.IsLFSAvailable() {
		log.Debugf(
			"[git-lfs] executing hook: git lfs %s %s", r.HookName, strings.Join(r.GitArgs, " "),
		)
		out := bytes.NewBuffer(make([]byte, 0))
		err := r.executor.RawExecute(
			ctx,
			append(
				[]string{"git", "lfs", r.HookName},
				r.GitArgs...,
			),
			r.stdin,
			out,
		)

		output := strings.Trim(out.String(), "\n")
		if output != "" {
			log.Debug("[git-lfs] out: ", output)
		}
		if err != nil {
			log.Debug("[git-lfs] err: ", err)
		}

		if err == nil && output != "" {
			log.Info(output)
		}

		if err != nil && (requiredExists || configExists) {
			log.Warnf("git-lfs command failed: %s\n", output)
			return err
		}

		return nil
	}

	if requiredExists || configExists {
		log.Errorf(
			"This Repository requires Git LFS, but 'git-lfs' wasn't found.\n"+
				"Install 'git-lfs' or consider reviewing the files:\n"+
				"  - %s\n"+
				"  - %s\n",
			lfsRequiredFile, lfsConfigFile,
		)
		return errors.New("git-lfs is required")
	}

	return nil
}

func (r *Runner) preHook() {
	if !config.HookUsesStagedFiles(r.HookName) {
		return
	}

	partiallyStagedFiles, err := r.Repo.PartiallyStagedFiles()
	if err != nil {
		log.Warnf("Couldn't find partially staged files: %s\n", err)
		return
	}

	if len(partiallyStagedFiles) == 0 {
		return
	}

	log.Debug("[lefthook] saving partially staged files")

	r.partiallyStagedFiles = partiallyStagedFiles
	err = r.Repo.SaveUnstaged(r.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't save unstaged changes: %s\n", err)
		return
	}

	err = r.Repo.StashUnstaged()
	if err != nil {
		log.Warnf("Couldn't stash partially staged files: %s\n", err)
		return
	}

	err = r.Repo.HideUnstaged(r.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't hide unstaged files: %s\n", err)
		return
	}

	log.Debugf("[lefthook] hide partially staged files: %v\n", r.partiallyStagedFiles)
}

func (r *Runner) postHook() {
	if !config.HookUsesStagedFiles(r.HookName) {
		return
	}

	if err := r.Repo.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		return
	}

	if err := r.Repo.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
	}
}

func (r *Runner) runScripts(ctx context.Context, dir string) []Result {
	files, err := afero.ReadDir(r.Repo.Fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return nil
	}

	scripts := make([]string, 0, len(files))
	filesMap := make(map[string]os.FileInfo)
	for _, file := range files {
		filesMap[file.Name()] = file
		scripts = append(scripts, file.Name())
	}
	sortByPriority(scripts, r.Hook.Scripts)

	interactiveScripts := make([]os.FileInfo, 0)
	var wg sync.WaitGroup
	resChan := make(chan Result, len(r.Hook.Scripts))
	results := make([]Result, 0, len(r.Hook.Scripts))

	for _, name := range scripts {
		file := filesMap[name]

		if ctx.Err() != nil {
			return nil
		}

		script, ok := r.Hook.Scripts[file.Name()]
		if !ok {
			r.logSkip(file.Name(), "not specified in config file")
			continue
		}

		if r.failed.Load() && r.Hook.Piped {
			r.logSkip(file.Name(), "broken pipe")
			continue
		}

		if script.Interactive && !r.Hook.Piped {
			interactiveScripts = append(interactiveScripts, file)
			continue
		}

		path := filepath.Join(dir, file.Name())

		if r.Hook.Parallel {
			wg.Add(1)
			go func(script *config.Script, path string, file os.FileInfo, resChan chan Result) {
				defer wg.Done()
				resChan <- r.runScript(ctx, script, path, file)
			}(script, path, file, resChan)
		} else {
			results = append(results, r.runScript(ctx, script, path, file))
		}
	}

	wg.Wait()
	close(resChan)
	for result := range resChan {
		results = append(results, result)
	}

	for _, file := range interactiveScripts {
		if ctx.Err() != nil {
			return nil
		}

		script := r.Hook.Scripts[file.Name()]
		if r.failed.Load() {
			r.logSkip(file.Name(), "non-interactive scripts failed")
			continue
		}

		path := filepath.Join(dir, file.Name())
		results = append(results, r.runScript(ctx, script, path, file))
	}

	return results
}

func (r *Runner) runScript(ctx context.Context, script *config.Script, path string, file os.FileInfo) Result {
	command, err := r.prepareScript(script, path, file)
	if err != nil {
		r.logSkip(file.Name(), err.Error())

		var skipErr *skipError
		if errors.As(err, &skipErr) {
			return skipped(file.Name())
		}

		r.failed.Store(true)
		return failed(file.Name(), err.Error())
	}

	if script.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	ok := r.run(ctx, exec.Options{
		Name:        file.Name(),
		Root:        r.Repo.RootPath,
		Commands:    []string{command},
		Interactive: script.Interactive && !r.DisableTTY,
		UseStdin:    script.UseStdin,
		Env:         script.Env,
	}, r.Hook.Follow)

	if !ok {
		r.failed.Store(true)
		return failed(file.Name(), script.FailText)
	}

	result := succeeded(file.Name())

	if config.HookUsesStagedFiles(r.HookName) && script.StageFixed {
		files, err := r.Repo.StagedFiles()
		if err != nil {
			log.Warn("Couldn't stage fixed files:", err)
			return result
		}

		r.addStagedFiles(files)
	}

	return result
}

func (r *Runner) runCommands(ctx context.Context) []Result {
	commands := make([]string, 0, len(r.Hook.Commands))
	for name := range r.Hook.Commands {
		if len(r.RunOnlyCommands) == 0 || slices.Contains(r.RunOnlyCommands, name) {
			commands = append(commands, name)
		}
	}

	sortByPriority(commands, r.Hook.Commands)

	interactiveCommands := make([]string, 0)
	var wg sync.WaitGroup
	results := make([]Result, 0, len(r.Hook.Commands))
	resChan := make(chan Result, len(r.Hook.Commands))

	for _, name := range commands {
		if r.failed.Load() && r.Hook.Piped {
			r.logSkip(name, "broken pipe")
			continue
		}

		if r.Hook.Commands[name].Interactive && !r.Hook.Piped {
			interactiveCommands = append(interactiveCommands, name)
			continue
		}

		if r.Hook.Parallel {
			wg.Add(1)
			go func(name string, command *config.Command, resChan chan Result) {
				defer wg.Done()
				result := r.runCommand(ctx, name, command)
				resChan <- result
			}(name, r.Hook.Commands[name], resChan)
		} else {
			result := r.runCommand(ctx, name, r.Hook.Commands[name])
			results = append(results, result)
		}
	}

	wg.Wait()
	close(resChan)
	for result := range resChan {
		results = append(results, result)
	}

	for _, name := range interactiveCommands {
		if r.failed.Load() {
			r.logSkip(name, "non-interactive commands failed")
			continue
		}

		results = append(results, r.runCommand(ctx, name, r.Hook.Commands[name]))
	}

	return results
}

func (r *Runner) runCommand(ctx context.Context, name string, command *config.Command) Result {
	run, err := r.prepareCommand(name, command)
	if err != nil {
		r.logSkip(name, err.Error())

		var skipErr *skipError
		if errors.As(err, &skipErr) {
			return skipped(name)
		}

		r.failed.Store(true)
		return failed(name, err.Error())
	}

	if command.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	ok := r.run(ctx, exec.Options{
		Name:        name,
		Root:        filepath.Join(r.Repo.RootPath, command.Root),
		Commands:    run.commands,
		Interactive: command.Interactive && !r.DisableTTY,
		UseStdin:    command.UseStdin,
		Env:         command.Env,
	}, r.Hook.Follow)

	if !ok {
		r.failed.Store(true)
		return failed(name, command.FailText)
	}

	result := succeeded(name)

	if config.HookUsesStagedFiles(r.HookName) && command.StageFixed {
		files := run.files

		if len(files) == 0 {
			var err error
			files, err = r.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result
			}

			files = filters.Apply(r.Repo.Fs, command, files)
		}

		if len(command.Root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(command.Root, file)
			}
		}

		r.addStagedFiles(files)
	}

	return result
}

func (r *Runner) addStagedFiles(files []string) {
	if err := r.Repo.AddFiles(files); err != nil {
		log.Warn("Couldn't stage fixed files:", err)
	}
}

func (r *Runner) run(ctx context.Context, opts exec.Options, follow bool) bool {
	log.SetName(opts.Name)
	defer log.UnsetName(opts.Name)

	// If the command does not explicitly `use_stdin` no input will be provided.
	var in io.Reader = NewNullReader()
	if opts.UseStdin {
		in = r.stdin
	}

	if (follow || opts.Interactive) && r.LogSettings.LogExecution() {
		r.logExecute(opts.Name, nil, nil)

		var out io.Writer
		if r.LogSettings.LogExecutionOutput() {
			out = os.Stdout
		} else {
			out = io.Discard
		}

		err := r.executor.Execute(ctx, opts, in, out)

		return err == nil
	}

	out := bytes.NewBuffer(make([]byte, 0))

	err := r.executor.Execute(ctx, opts, in, out)

	r.logExecute(opts.Name, err, out)

	return err == nil
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

func (r *Runner) logSkip(name, reason string) {
	if !r.LogSettings.LogSkips() {
		return
	}

	log.Styled().
		WithLeftBorder(lipgloss.NormalBorder(), log.ColorCyan).
		WithPadding(execLogPadding).
		Info(
			log.Cyan(log.Bold(name)) + " " +
				log.Gray("(skip)") + " " +
				log.Yellow(reason),
		)
}

func (r *Runner) logExecute(name string, err error, out io.Reader) {
	if err == nil && !r.LogSettings.LogExecution() {
		return
	}

	var execLog string
	var color lipgloss.TerminalColor
	switch {
	case !r.LogSettings.LogExecutionInfo():
		execLog = ""
	case err != nil:
		execLog = log.Red(fmt.Sprintf("%s ❯ ", name))
		color = log.ColorRed
	default:
		execLog = log.Cyan(fmt.Sprintf("%s ❯ ", name))
		color = log.ColorCyan
	}

	if execLog != "" {
		log.Styled().
			WithLeftBorder(lipgloss.ThickBorder(), color).
			WithPadding(execLogPadding).
			Info(execLog)
		log.Info()
	}

	if err == nil && !r.LogSettings.LogExecutionOutput() {
		return
	}

	if out != nil {
		log.Info(out)
	}

	if err != nil {
		log.Infof("%s", err)
	}
}

// sortByPriority sorts the tags by preceding numbers if they occur and special priority if it is set.
// If the names starts with letter the command name will be sorted alphabetically.
// If there's a `priority` field defined for a command or script it will be used instead of alphanumeric sorting.
//
//	[]string{"1_command", "10command", "3 command", "command5"} // -> 1_command, 3 command, 10command, command5
func sortByPriority[E executable](tags []string, executables map[string]E) {
	sort.SliceStable(tags, func(i, j int) bool {
		exeI, okI := executables[tags[i]]
		exeJ, okJ := executables[tags[j]]

		if okI && exeI.ExecutionPriority() != 0 || okJ && exeJ.ExecutionPriority() != 0 {
			if !okI || exeI.ExecutionPriority() == 0 {
				return false
			}
			if !okJ || exeJ.ExecutionPriority() == 0 {
				return true
			}

			return exeI.ExecutionPriority() < exeJ.ExecutionPriority()
		}

		numEnds := -1
		for idx, ch := range tags[i] {
			if unicode.IsDigit(ch) {
				numEnds = idx
			} else {
				break
			}
		}
		if numEnds == -1 {
			return tags[i] < tags[j]
		}
		numI, err := strconv.Atoi(tags[i][:numEnds+1])
		if err != nil {
			return tags[i] < tags[j]
		}

		numEnds = -1
		for idx, ch := range tags[j] {
			if unicode.IsDigit(ch) {
				numEnds = idx
			} else {
				break
			}
		}
		if numEnds == -1 {
			return true
		}
		numJ, err := strconv.Atoi(tags[j][:numEnds+1])
		if err != nil {
			return true
		}

		return numI < numJ
	})
}
