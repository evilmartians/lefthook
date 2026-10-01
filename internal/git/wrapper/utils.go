package wrapper

import (
	"iter"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func trimmed(seq iter.Seq[string]) iter.Seq[string] {
	return func(yield func(path string) bool) {
		next, stop := iter.Pull(seq)
		defer stop()

		for {
			value, ok := next()
			if !ok {
				return
			}

			value = strings.TrimSpace(value)
			if len(value) == 0 {
				continue
			}

			if !yield(value) {
				return
			}
		}
	}

}

func unquoted(seq iter.Seq[string]) iter.Seq[string] {
	return func(yield func(path string) bool) {
		next, stop := iter.Pull(seq)
		defer stop()

		for {
			value, ok := next()
			if !ok {
				return
			}

			unquoted, err := strconv.Unquote(value)
			if err == nil {
				value = unquoted // errors swallowed: bad but acceptable for now
			}

			if !yield(value) {
				return
			}
		}
	}
}

func (w *Wrapper) selectFiles(seq iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		next, stop := iter.Pull(seq)
		defer stop()

		for {
			value, ok := next()
			if !ok {
				return
			}

			if !w.isFile(value) {
				continue
			}

			if !yield(value) {
				return
			}
		}
	}
}

func (w *Wrapper) isFile(path string) bool {
	if !strings.HasPrefix(path, w.cmd.root) {
		path = filepath.Join(w.cmd.root, path)
	}

	stat, err := w.fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		return false // error swallowed
	}

	return !stat.IsDir()
}
