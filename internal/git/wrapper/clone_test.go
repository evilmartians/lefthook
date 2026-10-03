package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_Clone(t *testing.T) {
	fs := afero.NewMemMapFs()
	errClone := errors.New("clone failed")

	for name, tt := range map[string]struct {
		args wrapper.CloneArgs
		out  cmdtest.Out
		err  error
	}{
		"without-ref": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"},
			out: cmdtest.Out{
				Command: "git -C /root clone --quiet --origin origin --depth 1 https://example.com/repo.git dest",
			},
		},
		"with-ref": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git", Ref: "v1.0.0"},
			out: cmdtest.Out{
				Command: "git -C /root clone --quiet --origin origin --depth 1 --branch v1.0.0 https://example.com/repo.git dest",
			},
		},
		"clone-fails": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"},
			out: cmdtest.Out{
				Command: "git -C /root clone --quiet --origin origin --depth 1 https://example.com/repo.git dest",
				Err:     errClone,
			},
			err: errClone,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, []cmdtest.Out{tt.out})
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.Clone(tt.args)

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Clone() error = %v, want %v", err, tt.err)
			}
		})
	}
}
