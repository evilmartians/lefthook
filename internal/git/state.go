package git

import (
	"os"
	"path/filepath"
)

type State string

const (
	NilState    State = ""
	MergeState  State = "merge"
	RebaseState State = "rebase"
)

func (r *Repository) State() State {
	if r.isMergeState() {
		return MergeState
	}
	if r.isRebaseState() {
		return RebaseState
	}
	return NilState
}

func (r *Repository) isMergeState() bool {
	if _, err := os.Stat(filepath.Join(r.gitPath, "MERGE_HEAD")); os.IsNotExist(err) {
		return false
	}
	return true
}

func (r *Repository) isRebaseState() bool {
	if _, mergeErr := os.Stat(filepath.Join(r.gitPath, "rebase-merge")); os.IsNotExist(mergeErr) {
		if _, applyErr := os.Stat(filepath.Join(r.gitPath, "rebase-apply")); os.IsNotExist(applyErr) {
			return false
		}
	}

	return true
}
