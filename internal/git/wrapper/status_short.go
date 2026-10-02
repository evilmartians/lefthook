package wrapper

import (
	"slices"
	"strings"
)

var cmdStatusShort = []string{
	"git", "status", "--short", "--porcelain", "-z",
}

type FileStatus struct {
	Path     string
	Index    rune
	Worktree rune
}

// See https://git-scm.com/docs/git-status#_short_format.
func (w *Wrapper) StatusShort() ([]FileStatus, error) {
	output, err := w.cmd.WithoutTrim().cmd(cmdStatusShort) // there should be only one line with -z
	if err != nil {
		return nil, err
	}

	results := make([]FileStatus, 0)

	// parse short NUL separated porcelain v1 status output.
	skip := false
	for item := range strings.SplitSeq(output, "\x00") {
		if skip {
			skip = false
			continue
		}

		rs := []rune(item)
		if len(rs) < 4 || rs[2] != ' ' { // two status characters, space, and a filename
			continue
		}

		if slices.ContainsFunc(rs[0:2], func(r rune) bool {
			return r == 'C' || r == 'R'
		}) {
			// Next item after a Copy or Rename one is expected to be the old name, which we ignore
			skip = true
		}

		results = append(results, FileStatus{
			Path:     string(rs[3:]),
			Index:    rs[0],
			Worktree: rs[1],
		})
	}

	return results, nil
}
