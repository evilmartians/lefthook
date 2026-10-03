package wrapper

import "strings"

func (w *Wrapper) HashObjects(paths []string) ([]string, error) {
	out, err := w.cmd.batchedCmd([]string{"git", "hash-object", "--"}, paths)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}
