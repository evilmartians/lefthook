package git

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/log"
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

func (r *Repository) State() State {
	return r.stateOnce()
}

func (r *Repository) state() State {
	var state State

	branch := r.branch()
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

func (r *Repository) branch() string {
	headFile := filepath.Join(r.GitPath, "HEAD")
	if _, err := r.Fs.Stat(headFile); os.IsNotExist(err) {
		return ""
	}

	file, err := r.Fs.Open(headFile)
	if err != nil {
		return ""
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Warnf("Could not close %s: %s", headFile, cErr)
		}
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
