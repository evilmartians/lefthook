package wrapper

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type State struct {
	Branch, State string
}

const (
	nilState    string = ""
	merge       string = "merge"
	mergeCommit string = "merge-commit"
	rebase      string = "rebase"
)

var (
	refBranchRegexp  = regexp.MustCompile(`^ref:\s*refs/heads/(.+)$`)
	cmdParentCommits = []string{"git", "show", "--no-patch", `--format="%P"`}
)

func (w *Wrapper) State() State {
	var state State

	branch := w.branch()
	if w.inMergeState() {
		state = State{
			Branch: branch,
			State:  merge,
		}
		return state
	}
	if w.inRebaseState() {
		state = State{
			Branch: branch,
			State:  rebase,
		}
		return state
	}
	if w.inMergeCommitState() {
		state = State{
			Branch: branch,
			State:  mergeCommit,
		}
		return state
	}

	state = State{
		Branch: branch,
		State:  nilState,
	}

	return state
}

func (w *Wrapper) branch() string {
	headFile := filepath.Join(w.cache.gitPath, "HEAD")
	if _, err := w.fs.Stat(headFile); os.IsNotExist(err) {
		return ""
	}

	file, err := w.fs.Open(headFile)
	if err != nil {
		return ""
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		match := refBranchRegexp.FindStringSubmatch(scanner.Text())

		if match != nil {
			return match[1]
		}
	}

	return ""
}

func (w *Wrapper) inMergeState() bool {
	if _, err := w.fs.Stat(filepath.Join(w.cache.gitPath, "MERGE_HEAD")); os.IsNotExist(err) {
		return false
	}
	return true
}

func (w *Wrapper) inRebaseState() bool {
	if _, mergeErr := w.fs.Stat(filepath.Join(w.cache.gitPath, "rebase-merge")); os.IsNotExist(mergeErr) {
		if _, applyErr := w.fs.Stat(filepath.Join(w.cache.gitPath, "rebase-apply")); os.IsNotExist(applyErr) {
			return false
		}
	}

	return true
}

func (w *Wrapper) inMergeCommitState() bool {
	parents, err := w.cmd.cmd(cmdParentCommits)
	if err != nil {
		return false
	}

	return strings.Contains(parents, " ")
}
