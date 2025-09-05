package run

import (
	"bytes"
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/exec"
	"github.com/evilmartians/lefthook/internal/run/filters"
	"github.com/evilmartians/lefthook/internal/run/jobs"
	"github.com/evilmartians/lefthook/internal/run/result"
	"github.com/evilmartians/lefthook/internal/run/utils"
	"github.com/evilmartians/lefthook/internal/system"
)

var ErrFailOnChanges = errors.New("files were modified by a hook, and fail_on_changes is enabled")

type Options struct {
	Repo            *git.Repository
	Hook            *config.Hook
	HookName        string
	GitArgs         []string
	DisableTTY      bool
	SkipLFS         bool
	Force           bool
	Exclude         []string
	Files           []string
	RunOnlyCommands []string
	RunOnlyJobs     []string
	RunOnlyTags     []string
	SourceDirs      []string
	Templates       map[string]string
	FailOnChanges   bool
}

// Controller responds for actual execution and handling the results.
type Controller struct {
	Options

	cachedStdin          io.Reader
	partiallyStagedFiles []string
	failed               atomic.Bool
	executor             exec.Executor
	cmd                  system.CommandWithContext

	didStash        bool
	changesetBefore map[string]string
}

func NewController(opts Options) *Controller {
	return &Controller{
		Options: opts,

		// Some hooks use STDIN for parsing data from Git. To allow multiple commands
		// and scripts access the same Git data STDIN is cached via CachedReadec.
		cachedStdin: utils.NewCachedReader(os.Stdin),
		executor:    exec.CommandExecutor{},
		cmd:         system.Cmd,
	}
}

type executable interface {
	*config.Command | *config.Script
	ExecutionPriority() int
}

// RunAll runs scripts and commands.
// LFS hook is executed at first if needed.
func (c *Controller) RunAll(ctx context.Context) ([]result.Result, error) {
	results := make([]result.Result, 0, len(c.Hook.Commands)+len(c.Hook.Scripts))

	if config.NewSkipChecker(system.Cmd).Check(c.Repo.State, c.Hook.Skip, c.Hook.Only) {
		log.Skip(c.HookName, "hook setting")
		return results, nil
	}

	if err := c.runLFSHook(ctx); err != nil {
		return results, err
	}

	if !c.DisableTTY && !c.Hook.Follow {
		log.StartSpinner()
		defer log.StopSpinner()
	}

	scriptDirs := make([]string, 0, len(c.SourceDirs))
	for _, sourceDir := range c.SourceDirs {
		scriptDirs = append(scriptDirs, filepath.Join(
			sourceDir, c.HookName,
		))
	}

	c.preHook()

	results = append(results, c.runJobs(ctx)...)

	for _, dir := range scriptDirs {
		results = append(results, c.runScripts(ctx, dir)...)
	}

	results = append(results, c.runCommands(ctx)...)

	if err := c.postHook(); err != nil {
		return results, err
	}

	return results, nil
}

func (c *Controller) runLFSHook(ctx context.Context) error {
	if c.SkipLFS {
		return nil
	}

	if !git.IsLFSHook(c.HookName) {
		return nil
	}

	// Skip running git-lfs for pre-push hook when triggered manually
	if len(c.GitArgs) == 0 && c.HookName == "pre-push" {
		return nil
	}

	lfsRequiredFile := filepath.Join(c.Repo.RootPath, git.LFSRequiredFile)
	lfsConfigFile := filepath.Join(c.Repo.RootPath, git.LFSConfigFile)

	requiredExists, err := afero.Exists(c.Repo.Fs, lfsRequiredFile)
	if err != nil {
		return err
	}
	configExists, err := afero.Exists(c.Repo.Fs, lfsConfigFile)
	if err != nil {
		return err
	}

	if !git.IsLFSAvailable() {
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

	log.Debugf(
		"[git-lfs] executing hook: git lfs %s %s", c.HookName, strings.Join(c.GitArgs, " "),
	)
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	err = c.cmd.RunWithContext(
		ctx,
		append(
			[]string{"git", "lfs", c.HookName},
			c.GitArgs...,
		),
		"",
		c.cachedStdin,
		out,
		errOut,
	)

	outString := strings.Trim(out.String(), "\n")
	if outString != "" {
		log.Debug("[git-lfs] stdout: ", outString)
	}
	errString := strings.Trim(errOut.String(), "\n")
	if errString != "" {
		log.Debug("[git-lfs] stderr: ", errString)
	}
	if err != nil {
		log.Debug("[git-lfs] error:  ", err)
	}

	if err == nil && outString != "" {
		log.Info("[git-lfs] stdout: ", outString)
	}

	if err != nil && (requiredExists || configExists) {
		log.Warn("git-lfs command failed")
		if len(outString) > 0 {
			log.Warn("[git-lfs] stdout: ", outString)
		}
		if len(errString) > 0 {
			log.Warn("[git-lfs] stderr: ", errString)
		}
		return err
	}

	return nil
}

func (c *Controller) preHook() {
	if c.FailOnChanges {
		changeset, err := c.Repo.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
		} else {
			c.changesetBefore = changeset
		}
	}

	if !config.HookUsesStagedFiles(c.HookName) {
		return
	}

	partiallyStagedFiles, err := c.Repo.PartiallyStagedFiles()
	if err != nil {
		log.Warnf("Couldn't find partially staged files: %s\n", err)
		return
	}

	if len(partiallyStagedFiles) == 0 {
		return
	}

	c.didStash = true

	log.Debug("[lefthook] saving partially staged files")

	c.partiallyStagedFiles = partiallyStagedFiles
	err = c.Repo.SaveUnstaged(c.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't save unstaged changes: %s\n", err)
		return
	}

	err = c.Repo.StashUnstaged()
	if err != nil {
		log.Warnf("Couldn't stash partially staged files: %s\n", err)
		return
	}

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("hide partially staged files: ", c.partiallyStagedFiles).
		Log()

	err = c.Repo.HideUnstaged(c.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't hide unstaged files: %s\n", err)
		return
	}
}

func (c *Controller) postHook() error {
	if c.FailOnChanges {
		changesetAfter, err := c.Repo.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
		}
		if !maps.Equal(c.changesetBefore, changesetAfter) {
			return ErrFailOnChanges
		}
	}

	if !c.didStash {
		return
	}

	if err := c.Repo.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		return nil
	}

	if err := c.Repo.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
		return nil
	}

	return nil
}

func (c *Controller) runScripts(ctx context.Context, dir string) []result.Result {
	files, err := afero.ReadDir(c.Repo.Fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return nil
	}

	scripts := make([]string, 0, len(files))
	filesMap := make(map[string]os.FileInfo)
	for _, file := range files {
		filesMap[file.Name()] = file
		scripts = append(scripts, file.Name())
	}
	sortByPriority(scripts, c.Hook.Scripts)

	interactiveScripts := make([]os.FileInfo, 0)
	var wg sync.WaitGroup
	resChan := make(chan result.Result, len(c.Hook.Scripts))
	results := make([]result.Result, 0, len(c.Hook.Scripts))

	for _, name := range scripts {
		file := filesMap[name]

		if ctx.Err() != nil {
			return nil
		}

		script, ok := c.Hook.Scripts[file.Name()]
		if !ok {
			log.Skip(file.Name(), "not specified in config file")
			continue
		}

		if c.failed.Load() && c.Hook.Piped {
			log.Skip(file.Name(), "broken pipe")
			continue
		}

		if script.Interactive && !c.Hook.Piped {
			interactiveScripts = append(interactiveScripts, file)
			continue
		}

		if c.Hook.Parallel {
			wg.Add(1)
			go func(script *config.Script, file os.FileInfo, resChan chan result.Result) {
				defer wg.Done()
				resChan <- c.runScript(ctx, script, file)
			}(script, file, resChan)
		} else {
			results = append(results, c.runScript(ctx, script, file))
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

		script := c.Hook.Scripts[file.Name()]
		if c.failed.Load() {
			log.Skip(file.Name(), "non-interactive scripts failed")
			continue
		}

		results = append(results, c.runScript(ctx, script, file))
	}

	return results
}

func (c *Controller) runScript(ctx context.Context, script *config.Script, file os.FileInfo) result.Result {
	startTime := time.Now()

	job, err := jobs.Build(&jobs.Params{
		Name:   file.Name(),
		Runner: script.Runner,
		Script: file.Name(),
		Tags:   script.Tags,
		Only:   script.Only,
		Skip:   script.Skip,
	}, &jobs.Settings{
		Repo:       c.Repo,
		Hook:       c.Hook,
		HookName:   c.HookName,
		ForceFiles: c.Files,
		Force:      c.Force,
		GitArgs:    c.GitArgs,
		SourceDirs: c.SourceDirs,
		OnlyTags:   c.RunOnlyTags,
	})
	if err != nil {
		log.Skip(file.Name(), err.Error())

		var skipErr jobs.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(file.Name())
		}

		c.failed.Store(true)
		return result.Failure(file.Name(), err.Error(), time.Since(startTime))
	}

	if script.Interactive && !c.DisableTTY && !c.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	ok := exec.Run(ctx, c.executor, &exec.RunOptions{
		Exec: exec.Options{
			Name:        file.Name(),
			Root:        c.Repo.RootPath,
			Commands:    job.Execs,
			Interactive: script.Interactive && !c.DisableTTY,
			UseStdin:    script.UseStdin,
			Env:         script.Env,
		},
		Follow:      c.Hook.Follow,
		CachedStdin: c.cachedStdin,
	})

	executionTime := time.Since(startTime)

	if !ok {
		c.failed.Store(true)
		return result.Failure(file.Name(), script.FailText, executionTime)
	}

	result := result.Success(file.Name(), executionTime)

	if config.HookUsesStagedFiles(c.HookName) && script.StageFixed {
		files, err := c.Repo.StagedFiles()
		if err != nil {
			log.Warn("Couldn't stage fixed files:", err)
			return result
		}

		c.addStagedFiles(files)
	}

	return result
}

func (c *Controller) runCommands(ctx context.Context) []result.Result {
	commands := make([]string, 0, len(c.Hook.Commands))
	for name, command := range c.Hook.Commands {
		if len(c.RunOnlyCommands) != 0 && !slices.Contains(c.RunOnlyCommands, name) {
			continue
		}

		if len(c.RunOnlyTags) != 0 && !utils.Intersect(c.RunOnlyTags, command.Tags) {
			continue
		}

		commands = append(commands, name)
	}

	sortByPriority(commands, c.Hook.Commands)

	interactiveCommands := make([]string, 0)
	var wg sync.WaitGroup
	results := make([]result.Result, 0, len(c.Hook.Commands))
	resChan := make(chan result.Result, len(c.Hook.Commands))

	for _, name := range commands {
		if c.failed.Load() && c.Hook.Piped {
			log.Skip(name, "broken pipe")
			continue
		}

		if c.Hook.Commands[name].Interactive && !c.Hook.Piped {
			interactiveCommands = append(interactiveCommands, name)
			continue
		}

		if c.Hook.Parallel {
			wg.Add(1)
			go func(name string, command *config.Command, resChan chan result.Result) {
				defer wg.Done()
				result := c.runCommand(ctx, name, command)
				resChan <- result
			}(name, c.Hook.Commands[name], resChan)
		} else {
			result := c.runCommand(ctx, name, c.Hook.Commands[name])
			results = append(results, result)
		}
	}

	wg.Wait()
	close(resChan)
	for result := range resChan {
		results = append(results, result)
	}

	for _, name := range interactiveCommands {
		if c.failed.Load() {
			log.Skip(name, "non-interactive commands failed")
			continue
		}

		results = append(results, c.runCommand(ctx, name, c.Hook.Commands[name]))
	}

	return results
}

func (c *Controller) runCommand(ctx context.Context, name string, command *config.Command) result.Result {
	startTime := time.Now()
	exclude := command.Exclude
	switch list := exclude.(type) {
	case string:
		// Can't merge with regexp exclude
	case []interface{}:
		for _, e := range c.Exclude {
			list = append(list, e)
		}
		exclude = list
	default:
		// In case it's nil – simply replace
		excludeList := make([]interface{}, len(c.Exclude))
		for i, e := range c.Exclude {
			excludeList[i] = e
		}
		exclude = excludeList
	}

	job, err := jobs.Build(&jobs.Params{
		Name:      name,
		Run:       command.Run,
		Root:      command.Root,
		Glob:      command.Glob,
		Files:     command.Files,
		FileTypes: command.FileTypes,
		Tags:      command.Tags,
		Exclude:   exclude,
		Only:      command.Only,
		Skip:      command.Skip,
	}, &jobs.Settings{
		Repo:       c.Repo,
		Hook:       c.Hook,
		HookName:   c.HookName,
		ForceFiles: c.Files,
		Force:      c.Force,
		GitArgs:    c.GitArgs,
		Templates:  c.Templates,
	})
	if err != nil {
		log.Skip(name, err.Error())

		var skipErr jobs.SkipError
		if errors.As(err, &skipErr) {
			return result.Skip(name)
		}

		c.failed.Store(true)
		return result.Failure(name, err.Error(), time.Since(startTime))
	}

	if command.Interactive && !c.DisableTTY && !c.Hook.Follow {
		log.StopSpinner()
		defer log.StartSpinner()
	}

	ok := exec.Run(ctx, c.executor, &exec.RunOptions{
		Exec: exec.Options{
			Name:        name,
			Root:        filepath.Join(c.Repo.RootPath, command.Root),
			Commands:    job.Execs,
			Interactive: command.Interactive && !c.DisableTTY,
			UseStdin:    command.UseStdin,
			Env:         command.Env,
		},
		Follow:      c.Hook.Follow,
		CachedStdin: c.cachedStdin,
	})

	executionTime := time.Since(startTime)

	if !ok {
		c.failed.Store(true)
		return result.Failure(name, command.FailText, executionTime)
	}

	result := result.Success(name, executionTime)

	if config.HookUsesStagedFiles(c.HookName) && command.StageFixed {
		files := job.Files

		if len(files) == 0 {
			var err error
			files, err = c.Repo.StagedFiles()
			if err != nil {
				log.Warn("Couldn't stage fixed files:", err)
				return result
			}

			files = filters.Apply(c.Repo.Fs, files, filters.Params{
				Glob:      command.Glob,
				Root:      command.Root,
				Exclude:   exclude,
				FileTypes: command.FileTypes,
			})
		}

		if len(command.Root) > 0 {
			for i, file := range files {
				files[i] = filepath.Join(command.Root, file)
			}
		}

		c.addStagedFiles(files)
	}

	return result
}

func (c *Controller) addStagedFiles(files []string) {
	if err := c.Repo.AddFiles(files); err != nil {
		log.Warn("Couldn't stage fixed files:", err)
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
