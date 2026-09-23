package runner

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

type countingReader struct {
	io.Reader
	reads int
}

func (r *countingReader) Read(p []byte) (int, error) {
	r.reads++
	return r.Reader.Read(p)
}

func TestNewStdin(t *testing.T) {
	const input = "Some example string\nMultiline"

	for name, tt := range map[string]struct {
		readers    int
		concurrent bool
	}{
		"no readers":         {readers: 0},
		"one reader":         {readers: 1},
		"many readers":       {readers: 5},
		"concurrent readers": {readers: 5, concurrent: true},
	} {
		t.Run(name, func(t *testing.T) {
			in := &countingReader{Reader: strings.NewReader(input)}
			stdin := newStdin(in, loggertest.NewExecution())

			read := func() {
				data, err := io.ReadAll(stdin())
				assert.NoError(t, err)
				assert.Equal(t, input, string(data))
			}

			var wg sync.WaitGroup
			for range tt.readers {
				if tt.concurrent {
					wg.Go(read)
				} else {
					read()
				}
			}
			wg.Wait()

			if tt.readers == 0 {
				assert.Zero(t, in.reads, "STDIN must not be read when no job uses it")
			}
		})
	}
}
