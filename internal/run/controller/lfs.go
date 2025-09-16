package controller

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
)

func (c *Controller) runLFSHook(ctx context.Context, hookName string, args []string) error {
	if !git.IsLFSHook(hookName) {
		return nil
	}

	// Skip running git-lfs for pre-push hook when triggered manually
	if len(args) == 0 && hookName == "pre-push" {
		return nil
	}

	lfsRequiredFile := filepath.Join(c.git.RootPath, git.LFSRequiredFile)
	lfsConfigFile := filepath.Join(c.git.RootPath, git.LFSConfigFile)

	requiredExists, err := afero.Exists(c.git.Fs, lfsRequiredFile)
	if err != nil {
		return err
	}
	configExists, err := afero.Exists(c.git.Fs, lfsConfigFile)
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
		"[git-lfs] executing hook: git lfs %s %s", hookName, strings.Join(args, " "),
	)
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	err = c.cmd.RunWithContext(
		ctx,
		append(
			[]string{"git", "lfs", hookName},
			args...,
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
