package git

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
	Nil         string = ""
	Merge       string = "merge"
	MergeCommit string = "merge-commit"
	Rebase      string = "rebase"
)

var (
	refBranchRegexp  = regexp.MustCompile(`^ref:\s*refs/heads/(.+)$`)
	cmdParentCommits = []string{"git", "show", "--no-patch", `--format="%P"`}
)

var (
	state            State
	stateInitialized bool
)

func ResetState() {
	stateInitialized = false
}

func (r *Repository) State() State {
	if stateInitialized {
		return state
	}

	stateInitialized = true
	branch := r.Branch()
	if r.inMergeState() {
		state = State{
			Branch: branch,
			State:  Merge,
		}
		return state
	}
	if r.inRebaseState() {
		state = State{
			Branch: branch,
			State:  Rebase,
		}
		return state
	}
	if r.inMergeCommitState() {
		state = State{
			Branch: branch,
			State:  MergeCommit,
		}
		return state
	}

	state = State{
		Branch: branch,
		State:  Nil,
	}
	return state
}

func (r *Repository) Branch() string {
	headFile := filepath.Join(r.GitPath, "HEAD")
	if _, err := r.Fs.Stat(headFile); os.IsNotExist(err) {
		return ""
	}

	file, err := r.Fs.Open(headFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		match := refBranchRegexp.FindStringSubmatch(scanner.Text())

		if len(match) > 1 {
			return match[1]
		}
	}

	return ""
}

func (r *Repository) inMergeState() bool {
	if _, err := r.Fs.Stat(filepath.Join(r.GitPath, "MERGE_HEAD")); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *Repository) inRebaseState() bool {
	if _, mergeErr := r.Fs.Stat(filepath.Join(r.GitPath, "rebase-merge")); os.IsNotExist(mergeErr) {
		if _, applyErr := r.Fs.Stat(filepath.Join(r.GitPath, "rebase-apply")); os.IsNotExist(applyErr) {
			return false
		}
	}

	return true
}

func (r *Repository) inMergeCommitState() bool {
	parents, err := r.Git.Cmd(cmdParentCommits)
	if err != nil {
		return false
	}

	return strings.Contains(parents, " ")
}
