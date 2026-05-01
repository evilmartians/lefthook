package controller

import (
	"errors"
	"fmt"
	"maps"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
)

type FailOnChangesError struct {
	changedFiles []string
}

func (e *FailOnChangesError) Error() string {
	return "files were modified by a hook, and fail_on_changes is enabled"
}

type guard struct {
	git *git.Repository

	stashUnstagedChanges bool
	failOnChanges        bool
	failOnChangesDiff    bool
}

func newGuard(repo *git.Repository, stashUnstagedChanges bool, failOnChanges bool, failOnChangesDiff bool) *guard {
	return &guard{
		git:                  repo,
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
		log.Warnf("Couldn't find partially staged files: %s\n", err)
		return err
	}

	if len(partiallyStagedFiles) == 0 {
		return fn()
	}

	log.Debug("[lefthook] saving partially staged files")

	if err := g.git.SaveUnstaged(partiallyStagedFiles); err != nil {
		log.Warnf("Couldn't save unstaged changes: %s\n", err)
		return err
	}

	if err := g.git.StashUnstaged(); err != nil {
		log.Warnf("Couldn't stash partially staged files: %s\n", err)
		return err
	}

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("hide partially staged files: ", partiallyStagedFiles).
		Log()

	if err := g.git.RevertUnstagedChanges(partiallyStagedFiles); err != nil {
		log.Warnf("Couldn't hide unstaged files: %s\n", err)
		return err
	}

	wrappedErr := fn()

	var failOnChangesErr *FailOnChangesError
	if errors.As(wrappedErr, &failOnChangesErr) {
		if err := g.git.RevertUnstagedChanges(failOnChangesErr.changedFiles); err != nil {
			log.Warnf("Couldn't hide unstaged files: %s\n", err)
			return wrappedErr
		}
	}

	if err := g.git.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		return wrappedErr
	}

	if err := g.git.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
		return wrappedErr
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
		return fmt.Errorf("couldn't calculate changeset: %w", err)
	}

	fn()

	changesetAfter, err := g.git.Changeset()
	if err != nil {
		return fmt.Errorf("couldn't calculate changeset: %w", err)
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
