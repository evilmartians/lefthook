package run

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
	"strings"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/lefthook/run/exec"
	"github.com/evilmartians/lefthook/internal/lefthook/run/filter"
	"github.com/evilmartians/lefthook/internal/log"
)

type status int8

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
	ResultChan      chan Result
	SkipSettings    log.SkipSettings
	DisableTTY      bool
	AllFiles        bool
	Files           []string
	RunOnlyCommands []string
}

// Runner responds for actual execution and handling the results.
type Runner struct {
	Options

	partiallyStagedFiles []string
	failed               atomic.Bool
	executor             exec.Executor
}

func NewRunner(opts Options) *Runner {
	return &Runner{
		Options:  opts,
		executor: exec.CommandExecutor{},
	}
}

// RunAll runs scripts and commands.
// LFS hook is executed at first if needed.
func (r *Runner) RunAll(ctx context.Context, sourceDirs []string) {
	if err := r.runLFSHook(ctx); err != nil {
		log.Error(err)
	}

	if r.Hook.DoSkip(r.Repo.State()) {
		r.logSkip(r.HookName, "hook setting")
		return
	}

	if !r.DisableTTY && !r.Hook.Follow {
		log.StartSpinner()
		defer log.StopSpinner()
	}

	scriptDirs := make([]string, len(sourceDirs))
	for _, sourceDir := range sourceDirs {
		scriptDirs = append(scriptDirs, filepath.Join(
			sourceDir, r.HookName,
		))
	}

	r.preHook()

	for _, dir := range scriptDirs {
		r.runScripts(ctx, dir)
	}

	r.runCommands(ctx)

	r.postHook()
}

func (r *Runner) fail(name string, err error) {
	r.ResultChan <- resultFail(name, err.Error())
	r.failed.Store(true)
}

func (r *Runner) success(name string) {
	r.ResultChan <- resultSuccess(name)
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
		log.Warnf("Couldn't restore hidden unstaged files: %s\n", err)
		return
	}

	if err := r.Repo.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
	}
}

func (r *Runner) runScripts(ctx context.Context, dir string) {
	files, err := afero.ReadDir(r.Repo.Fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return
	}

	interactiveScripts := make([]os.FileInfo, 0)
	var wg sync.WaitGroup

	for _, file := range files {
		if ctx.Err() != nil {
			return
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

		if script.Interactive {
			interactiveScripts = append(interactiveScripts, file)
			continue
		}

		path := filepath.Join(dir, file.Name())

		if r.Hook.Parallel {
			wg.Add(1)
			go func(script *config.Script, path string, file os.FileInfo) {
				defer wg.Done()
				r.runScript(ctx, script, path, file)
			}(script, path, file)
		} else {
			r.runScript(ctx, script, path, file)
		}
	}

	wg.Wait()

	for _, file := range interactiveScripts {
		if ctx.Err() != nil {
			return
		}

		script := r.Hook.Scripts[file.Name()]
		if r.failed.Load() {
			r.logSkip(file.Name(), "non-interactive scripts failed")
			continue
		}

		path := filepath.Join(dir, file.Name())
		r.runScript(ctx, script, path, file)
	}
}

func (r *Runner) runScript(ctx context.Context, script *config.Script, path string, file os.FileInfo) {
	command, err := r.prepareScript(script, path, file)
	if err != nil {
		r.logSkip(file.Name(), err.Error())
		return
	}

	if script.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	finished := r.run(ctx, exec.Options{
		Name:        file.Name(),
		Root:        r.Repo.RootPath,
		Commands:    []string{command},
		FailText:    script.FailText,
		Interactive: script.Interactive && !r.DisableTTY,
		UseStdin:    script.UseStdin,
		Env:         script.Env,
	}, r.Hook.Follow)

	if finished && config.HookUsesStagedFiles(r.HookName) && script.StageFixed {
		files, err := r.Repo.StagedFiles()
		if err != nil {
			log.Warn("Couldn't stage fixed files:", err)
			return
		}

		r.addStagedFiles(files)
	}
}

func (r *Runner) runCommands(ctx context.Context) {
	commands := make([]string, 0, len(r.Hook.Commands))
	for name := range r.Hook.Commands {
		if len(r.RunOnlyCommands) == 0 || slices.Contains(r.RunOnlyCommands, name) {
			commands = append(commands, name)
		}
	}

	sort.Strings(commands)

	interactiveCommands := make([]string, 0)
	var wg sync.WaitGroup

	for _, name := range commands {
		if r.failed.Load() && r.Hook.Piped {
			r.logSkip(name, "broken pipe")
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
				r.runCommand(ctx, name, command)
			}(name, r.Hook.Commands[name])
		} else {
			r.runCommand(ctx, name, r.Hook.Commands[name])
		}
	}

	wg.Wait()

	for _, name := range interactiveCommands {
		if r.failed.Load() {
			r.logSkip(name, "non-interactive commands failed")
			continue
		}

		r.runCommand(ctx, name, r.Hook.Commands[name])
	}
}

func (r *Runner) runCommand(ctx context.Context, name string, command *config.Command) {
	run, err := r.prepareCommand(name, command)
	if err != nil {
		r.logSkip(name, err.Error())
		return
	}

	if command.Interactive && !r.DisableTTY && !r.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	finished := r.run(ctx, exec.Options{
		Name:        name,
		Root:        filepath.Join(r.Repo.RootPath, command.Root),
		Commands:    run.commands,
		FailText:    command.FailText,
		Interactive: command.Interactive && !r.DisableTTY,
		UseStdin:    command.UseStdin,
		Env:         command.Env,
	}, r.Hook.Follow)

	if finished && config.HookUsesStagedFiles(r.HookName) && command.StageFixed {
		files := run.files

		if len(files) == 0 {
			var err error
			files, err = r.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return
			}

			files = filter.Apply(command, files)
		}

		if len(command.Root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(command.Root, file)
			}
		}

		r.addStagedFiles(files)
	}
}

func (r *Runner) addStagedFiles(files []string) {
	if err := r.Repo.AddFiles(files); err != nil {
		log.Warn("Couldn't stage fixed files:", err)
	}
}

func (r *Runner) run(ctx context.Context, opts exec.Options, follow bool) bool {
	log.SetName(opts.Name)
	defer log.UnsetName(opts.Name)

	if (follow || opts.Interactive) && !r.SkipSettings.SkipExecution() {
		r.logExecute(opts.Name, nil, nil)

		var out io.Writer
		if r.SkipSettings.SkipExecutionOutput() {
			out = io.Discard
		} else {
			out = os.Stdout
		}

		err := r.executor.Execute(ctx, opts, out)
		if err != nil {
			r.fail(opts.Name, errors.New(opts.FailText))
		} else {
			r.success(opts.Name)
		}

		return err == nil
	}

	out := bytes.NewBuffer(make([]byte, 0))
	err := r.executor.Execute(ctx, opts, out)

	if err != nil {
		r.fail(opts.Name, errors.New(opts.FailText))
	} else {
		r.success(opts.Name)
	}

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
	if r.SkipSettings.SkipSkips() {
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
	if err == nil && r.SkipSettings.SkipExecution() {
		return
	}

	var execLog string
	var color lipgloss.TerminalColor
	switch {
	case r.SkipSettings.SkipExecutionInfo():
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

	if err == nil && r.SkipSettings.SkipExecutionOutput() {
		return
	}

	if out != nil {
		log.Info(out)
	}

	if err != nil {
		log.Infof("%s", err)
	}
}
