package controller

import (
	"errors"
	"fmt"
	"maps"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/logger"
)

var errRestorationConflict = errors.New("conflict while merging unstaged changes")

type FailOnChangesError struct {
	changedFiles []string
}

func (e *FailOnChangesError) Error() string {
	return "files were modified by a hook, and fail_on_changes is enabled"
}

type guard struct {
	git    *git.Repo
	logger *logger.ExecutionLogger

	stashUnstagedChanges bool
	failOnChanges        bool
	failOnChangesDiff    bool
}

func newGuard(
	repo *git.Repo,
	logger *logger.ExecutionLogger,
	stashUnstagedChanges bool,
	failOnChanges bool,
	failOnChangesDiff bool,
) *guard {
	return &guard{
		git:                  repo,
		logger:               logger,
		stashUnstagedChanges: stashUnstagedChanges,
		failOnChanges:        failOnChanges,
		failOnChangesDiff:    failOnChangesDiff,
	}
}

func (g *guard) wrap(fn func()) error {
	if !g.failOnChanges && !g.stashUnstagedChanges {
		fn()

		return nil
	}

	return g.withHiddenUnstagedChanges(
		func() error {
			return g.withFailOnChanges(fn)
		},
	)
}

func (g *guard) withHiddenUnstagedChanges(fn func() error) error {
	if !g.stashUnstagedChanges {
		return fn()
	}

	partiallyStagedFiles, err := g.git.PartiallyStagedFiles()
	if err != nil {
		g.logger.Warnf("Failed to find partially staged files: %s\n", err)
		return err
	}

	if len(partiallyStagedFiles) == 0 {
		return fn()
	}

	g.logger.Debug("[lefthook] saving partially staged files")

	if err := g.git.SaveUnstagedChanges(partiallyStagedFiles); err != nil {
		g.logger.Warnf("Failed to save unstaged changes: %s\n", err)
		return err
	}

	logger.NewBuilder(g.logger).
		WithPrefix("[lefthook] ").
		WithLevel(logger.LevelDebug).
		WriteLines("hide partially staged files: ", partiallyStagedFiles).
		Log()

	if err := g.git.RevertUnstagedChanges(partiallyStagedFiles); err != nil {
		g.logger.Warnf("Failed to hide unstaged files: %s", err)
		return err
	}

	wrappedErr := fn()

	var failOnChangesErr *FailOnChangesError
	if errors.As(wrappedErr, &failOnChangesErr) {
		if err := g.git.RevertUnstagedChanges(failOnChangesErr.changedFiles); err != nil {
			g.logger.Warnf("Failed to revert file changes: %s", err)
			return wrappedErr
		}
	}

	if !g.git.CanRestoreUnstagedChanges() {
		if wrappedErr != nil {
			g.logger.Error("Error: ", wrappedErr)
		}
		wrappedErr = errRestorationConflict

		if err := g.git.RevertAllUnstagedChanges(); err != nil {
			g.logger.Warnf("Failed to restore initial worktree state: %s", err)
			return err
		}

		logger.NewBuilder(g.logger).
			WithLevel(logger.LevelWarn).
			WriteLines("", "Unable to restore previously hidden unstaged changes.").
			WriteLines("", "This may happen when changes introduced by the hook conflict with your unstaged changes.").
			WriteLines("", "Stage all changes with `git add -A` and try again.").
			Log()
	}

	if err := g.git.RestoreUnstagedChanges(); err != nil {
		g.logger.Warnf("Failed to restore unstaged files: %s", err)
		return err
	}

	return wrappedErr
}

func (g *guard) withFailOnChanges(fn func()) error {
	if !g.failOnChanges {
		fn()
		return nil
	}

	changesetBefore, err := g.git.Changeset()
	if err != nil {
		return fmt.Errorf("changeset calculation failed: %w", err)
	}

	fn()

	changesetAfter, err := g.git.Changeset()
	if err != nil {
		return fmt.Errorf("changeset calculation failed: %w", err)
	}
	if !maps.Equal(changesetBefore, changesetAfter) {
		changedFiles := g.printDiff(changesetBefore, changesetAfter)
		return &FailOnChangesError{changedFiles: changedFiles}
	}

	return nil
}

func (g *guard) printDiff(changesetBefore, changesetAfter map[string]string) []string {
	changedFiles := g.getChangedFiles(changesetBefore, changesetAfter)

	if g.failOnChangesDiff && len(changedFiles) > 0 {
		g.git.PrintDiff(changedFiles)
	}

	return changedFiles
}

func (g *guard) getChangedFiles(changesetBefore, changesetAfter map[string]string) []string {
	changed := make([]string, 0, len(changesetBefore))
	for f, hashBefore := range changesetBefore {
		if hashAfter, ok := changesetAfter[f]; !ok || hashBefore != hashAfter {
			changed = append(changed, f)
		}
	}

	for f := range changesetAfter {
		if _, ok := changesetBefore[f]; !ok {
			changed = append(changed, f)
		}
	}

	return changed
}
