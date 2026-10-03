package wrapper_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_HashObjects(t *testing.T) {
	fs := afero.NewMemMapFs()
	errHash := errors.New("hash-object failed")

	for name, tt := range map[string]struct {
		out  cmdtest.Out
		want []string
		err  error
	}{
		"returns-hashes": {
			out:  cmdtest.Out{Command: "git hash-object -- a b", Output: "hash-a\nhash-b\n"},
			want: []string{"hash-a", "hash-b"},
		},
		"hash-object-fails": {
			out: cmdtest.Out{Command: "git hash-object -- a b", Err: errHash},
			err: errHash,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			result, err := w.HashObjects([]string{"a", "b"})

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.HashObjects() error = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.HashObjects() = %v, want %v", result, tt.want)
			}
		})
	}
}
