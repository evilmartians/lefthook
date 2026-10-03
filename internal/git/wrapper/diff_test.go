package wrapper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_Diff(t *testing.T) {
	fs := afero.NewMemMapFs()

	for name, tt := range map[string]struct {
		colors  bool
		files   []string
		command string
	}{
		"colors-enabled": {
			colors:  true,
			files:   []string{"file2", "file1"},
			command: "git diff --color -- file2 file1",
		},
		"colors-disabled": {
			colors:  false,
			files:   []string{"file2", "file1"},
			command: "git diff -- file2 file1",
		},
	} {
		t.Run(name, func(t *testing.T) {
			logger := loggertest.New()
			cmd := cmdtest.NewFakeCmd(t, []cmdtest.Out{{Command: tt.command, Output: "<anything>"}})

			w := wrapper.New(fs, cmd, logger)

			result, err := w.Diff(tt.files, tt.colors)
			wanted := "<anything>"

			if err != nil {
				t.Errorf("err = %v, want nil", err)
			}

			if !cmp.Equal(result, wanted) {
				t.Errorf("wrapper.Diff() = %v, wanted %v", result, wanted)
			}
		})
	}
}
