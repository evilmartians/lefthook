package runner

import (
	"bytes"
	"io"
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

// newStdin returns a function that gives a new reader over the same input on
// each call. Git passes hook data (e.g. refs for pre-push) via STDIN, and
// several jobs and git-lfs may need all of it. The input is read only once, on
// the first call, so hooks that do not use STDIN never block on it.
func newStdin(in io.Reader, log *logger.ExecutionLogger) func() io.Reader {
	read := sync.OnceValue(func() []byte {
		data, err := io.ReadAll(in)
		if err != nil {
			log.Warnf("Couldn't read STDIN: %s\n", err)
		}

		return data
	})

	return func() io.Reader {
		return bytes.NewReader(read())
	}
}
