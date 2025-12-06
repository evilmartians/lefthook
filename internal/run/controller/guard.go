package controller

import (
	"errors"
	"maps"
	"slices"

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

	g.before()
	call()
	return g.after()
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
}

func (g *guard) after() error {
	if g.failOnChanges {
		changesetAfter, err := g.git.Changeset()
		if err != nil {
			log.Warnf("Couldn't get changeset: %s\n", err)
		}
		if !maps.Equal(g.changesetBefore, changesetAfter) {
			g.changesetDiff(changesetAfter)
			return ErrFailOnChanges
		}
	}

	if !g.didStash {
		return nil
	}

	if err := g.git.RestoreUnstaged(); err != nil {
		log.Warnf("Couldn't restore unstaged files: %s\n", err)
		return nil
	}

	if err := g.git.DropUnstagedStash(); err != nil {
		log.Warnf("Couldn't remove unstaged files backup: %s\n", err)
		return nil
	}

	return nil
}

func (g *guard) changesetDiff(changesetAfter map[string]string) {
	if !g.failOnChangesDiff {
		return
	}
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
	if len(changed) == 0 {
		return
	}
	slices.Sort(changed)
	diffCmd := make([]string, 0, 4) //nolint:mnd // 3 or 4 elements
	diffCmd = append(diffCmd, "git", "diff")
	if log.Colorized() {
		diffCmd = append(diffCmd, "--color")
	}
	diffCmd = append(diffCmd, "--")
	if diff, err := g.git.Git.BatchedCmd(diffCmd, changed); err != nil {
		log.Warnf("Couldn't diff changed files: %s", err)
	} else {
		log.Println(diff)
	}
}
