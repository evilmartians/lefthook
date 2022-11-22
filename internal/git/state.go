package git

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

type State struct {
	Branch, Step string
}

const (
	NilStep    string = ""
	MergeStep  string = "merge"
	RebaseStep string = "rebase"
)

var refBranchRegexp = regexp.MustCompile(`^ref:\s*refs/heads/(.+)$`)

func (r *Repository) State() State {
	branch := r.Branch()
	if r.isMergeState() {
		return State{
			Branch: branch,
			Step:   MergeStep,
		}
	}
	if r.isRebaseState() {
		return State{
			Branch: branch,
			Step:   RebaseStep,
		}
	}
	return State{
		Branch: branch,
		Step:   NilStep,
	}
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

func (r *Repository) isMergeState() bool {
	if _, err := r.Fs.Stat(filepath.Join(r.GitPath, "MERGE_HEAD")); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *Repository) isRebaseState() bool {
	if _, mergeErr := r.Fs.Stat(filepath.Join(r.GitPath, "rebase-merge")); os.IsNotExist(mergeErr) {
		if _, applyErr := r.Fs.Stat(filepath.Join(r.GitPath, "rebase-apply")); os.IsNotExist(applyErr) {
			return false
		}
	}

	return true
}
