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
