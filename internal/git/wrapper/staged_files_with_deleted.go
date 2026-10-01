package wrapper

import "slices"

var cmdStagedFilesWithDeleted = []string{
	"git", "diff", "--name-only", "--cached", "--diff-filter=ACMRD",
}

func (w *Wrapper) StagedFilesWithDeleted() ([]string, error) {
	lines, err := w.cmd.CmdLines(cmdStagedFilesWithDeleted)
	if err != nil {
		return nil, err
	}

	return slices.Collect(unquoted(trimmed(slices.Values(lines)))), nil
}
