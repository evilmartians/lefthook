package controller

import (
	"errors"
	"maps"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
)

var ErrFailOnChanges = errors.New("files were modified by a hook, and fail_on_changes is enabled")

type guard struct {
	git *git.Repository

	stashUnstagedChanges bool
	failOnChanges        bool
	failOnChangesDiff    bool

	didStash             bool
	partiallyStagedFiles []string
	changesetBefore      map[string]string
}

func newGuard(repo *git.Repository, stashUnstagedChanges bool, failOnChanges bool, failOnChangesDiff bool) *guard {
	return &guard{
		git:                  repo,
		stashUnstagedChanges: stashUnstagedChanges,
		failOnChanges:        failOnChanges,
		failOnChangesDiff:    failOnChangesDiff,
	}
}

func (g *guard) wrap(call func()) error {
	if !g.failOnChanges && !g.stashUnstagedChanges {
		call()
		return nil
	}

	return g.withHiddenUnstagedChanges(
		g.withFailOnChanges(
			call,
		),
	)
	// g.before()
	//
	// call()
	//
	// return g.after()
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

	g.partiallyStagedFiles = partiallyStagedFiles
	err = g.git.SaveUnstaged(g.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't save unstaged changes: %s\n", err)
		return err
	}

	err = g.git.StashUnstaged()
	if err != nil {
		log.Warnf("Couldn't stash partially staged files: %s\n", err)
		return err
	}

	g.didStash = true

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("hide partially staged files: ", g.partiallyStagedFiles).
		Log()

	err = g.git.HideUnstaged(g.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't hide unstaged files: %s\n", err)
		return err
	}

	return fn()
}

func (g *guard) withFailOnChanges(fn func()) error {
	if !g.failOnChanges {
		fn()
		return nil
	}

	changeset, err := g.git.Changeset()
	if err != nil {
		log.Warnf("Couldn't get changeset: %s\n", err)
	} else {
		g.changesetBefore = changeset
	}

	fn()

	// Only get changeset if we need it for failOnChanges check
	var changesetAfter map[string]string
	var isFailingOnChanges bool
	var err error
	changesetAfter, err = g.git.Changeset()
	if err != nil {
		log.Warnf("Couldn't get changeset: %s\n", err)
		changesetAfter = make(map[string]string)
	}
	isFailingOnChanges = !maps.Equal(g.changesetBefore, changesetAfter)

	if !g.didStash {
		if isFailingOnChanges {
			g.printDiff(changesetAfter)
			return ErrFailOnChanges
		}
		return nil
	}

	if err := g.git.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		// If we can't restore the unstaged files, first roll back the changes
		// introduced by the hook before trying to restore unstaged files again
		// Get changeset only when needed for error recovery
		changesetAfter, err := g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
			changesetAfter = make(map[string]string)
		}
		changed := g.getChangedFiles(changesetAfter)

		log.Warnf("Couldn't restore unstaged files after hook changes, rolling back: %s\n", changed)
		err = g.git.HideUnstaged(changed)
		if err != nil {
			log.Warnf("Couldn't rollback hook changes: %s\n", err)
			return nil
		}

		// Retry restoring unstaged files after rolling back hook changes
		if retryErr := g.git.RestoreUnstaged(); retryErr != nil {
			log.Warnf("Couldn't restore unstaged files after rollback: %s\n", retryErr)
			return nil
		}
	}

	if err := g.git.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
		return nil
	}

	if isFailingOnChanges {
		g.printDiff(changesetAfter)
		return ErrFailOnChanges
	}

	return nil
}

func (g *guard) before() {
	if g.failOnChanges {
		changeset, err := g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
		} else {
			g.changesetBefore = changeset
		}
	}

	if !g.stashUnstagedChanges {
		return
	}

	partiallyStagedFiles, err := g.git.PartiallyStagedFiles()
	if err != nil {
		log.Warnf("Couldn't find partially staged files: %s\n", err)
		return
	}

	if len(partiallyStagedFiles) == 0 {
		return
	}

	log.Debug("[lefthook] saving partially staged files")

	g.partiallyStagedFiles = partiallyStagedFiles
	err = g.git.SaveUnstaged(g.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't save unstaged changes: %s\n", err)
		return
	}

	err = g.git.StashUnstaged()
	if err != nil {
		log.Warnf("Couldn't stash partially staged files: %s\n", err)
		return
	}

	g.didStash = true

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("hide partially staged files: ", g.partiallyStagedFiles).
		Log()

	err = g.git.HideUnstaged(g.partiallyStagedFiles)
	if err != nil {
		log.Warnf("Couldn't hide unstaged files: %s\n", err)
		return
	}

	// Capture changeset after stashing partially staged files, so we compare the same state
	// before and after running hooks
	if g.failOnChanges {
		changeset, err := g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
		} else {
			g.changesetBefore = changeset
		}
	}
}

func (g *guard) after() error {
	// Only get changeset if we need it for failOnChanges check
	var changesetAfter map[string]string
	var isFailingOnChanges bool

	if g.failOnChanges {
		var err error
		changesetAfter, err = g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
			changesetAfter = make(map[string]string)
		}
		isFailingOnChanges = !maps.Equal(g.changesetBefore, changesetAfter)
	}

	if !g.didStash {
		if isFailingOnChanges {
			g.printDiff(changesetAfter)
			return ErrFailOnChanges
		}
		return nil
	}

	if err := g.git.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		// If we can't restore the unstaged files, first roll back the changes
		// introduced by the hook before trying to restore unstaged files again
		// Get changeset only when needed for error recovery
		changesetAfter, err := g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
			changesetAfter = make(map[string]string)
		}
		changed := g.getChangedFiles(changesetAfter)

		log.Warnf("Couldn't restore unstaged files after hook changes, rolling back: %s\n", changed)
		err = g.git.HideUnstaged(changed)
		if err != nil {
			log.Warnf("Couldn't rollback hook changes: %s\n", err)
			return nil
		}

		// Retry restoring unstaged files after rolling back hook changes
		if retryErr := g.git.RestoreUnstaged(); retryErr != nil {
			log.Warnf("Couldn't restore unstaged files after rollback: %s\n", retryErr)
			return nil
		}
	}

	if err := g.git.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
		return nil
	}

	if isFailingOnChanges {
		g.printDiff(changesetAfter)
		return ErrFailOnChanges
	}

	return nil
}

func (g *guard) printDiff(changesetAfter map[string]string) {
	if !g.failOnChangesDiff {
		return
	}

	changed := g.getChangedFiles(changesetAfter)

	if len(changed) == 0 {
		return
	}

	g.git.PrintDiff(changed)
}

func (g *guard) getChangedFiles(changesetAfter map[string]string) []string {
	changed := make([]string, 0, len(g.changesetBefore))
	for f, hashBefore := range g.changesetBefore {
		if hashAfter, ok := changesetAfter[f]; !ok || hashBefore != hashAfter {
			changed = append(changed, f)
		}
	}

	for f := range changesetAfter {
		if _, ok := g.changesetBefore[f]; !ok {
			changed = append(changed, f)
		}
	}

	return changed
}
