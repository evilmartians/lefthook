package run

import (
	"bytes"
	"context"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/run/exec"
	"github.com/evilmartians/lefthook/internal/run/result"
	"github.com/evilmartians/lefthook/internal/run/utils"
	"github.com/evilmartians/lefthook/internal/system"
)

var ErrFailOnChanges = errors.New("files were modified by a hook, and fail_on_changes is enabled")

type Options struct {
	Repo          *git.Repository
	Hook          *config.Hook
	HookName      string
	GitArgs       []string
	DisableTTY    bool
	SkipLFS       bool
	Force         bool
	Exclude       []string
	Files         []string
	RunOnlyJobs   []string
	RunOnlyTags   []string
	SourceDirs    []string
	Templates     map[string]string
	FailOnChanges bool
}

// Controller responds for actual execution and handling the results.
type Controller struct {
	Options

	cachedStdin          io.Reader
	partiallyStagedFiles []string
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

	c.preHook()

	results = append(results, c.runJobs(ctx)...)

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
