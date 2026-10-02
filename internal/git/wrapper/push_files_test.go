package wrapper_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_PushFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := loggertest.New()

	for _, file := range []string{"a", "b", "c"} {
		if err := afero.WriteFile(fs, file, []byte("content"), 0o644); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}

	for name, tt := range map[string]struct {
		cmd  system.Command
		want []string
		err  error
	}{
		"@{push}": {
			cmd: cmdtest.NewOrdered(
				t,
				[]cmdtest.Out{
					{
						Command: "git diff --name-only HEAD @{push}",
						Output:  "a\nb\nc",
					},
				},
			),
			want: []string{"a", "b", "c"},
		},
		"head-branch": {
			cmd: cmdtest.NewOrdered(
				t,
				[]cmdtest.Out{
					{
						Command: "git diff --name-only HEAD @{push}",
						Err:     errors.New("oops"),
					},
					{
						Command: "git branch --remotes",
						Output:  "  origin/fix\n  origin/feat\n  origin/HEAD -> origin/main\n  origin/bug\n",
					},
					{
						Command: "git diff --name-only HEAD origin/main --",
						Output:  "a\nb\n",
					},
				},
			),
			want: []string{"a", "b"},
		},
		"ls-files": {
			cmd: cmdtest.NewOrdered(
				t,
				[]cmdtest.Out{
					{
						Command: "git diff --name-only HEAD @{push}",
						Err:     errors.New("oops"),
					},
					{
						Command: "git branch --remotes",
						Output:  "  origin/fix\n  origin/feat\n  origin/bug\n",
					},
					{
						Command: "git ls-tree -r --name-only HEAD",
						Output:  "a\n",
					},
				},
			),
			want: []string{"a"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := wrapper.New(fs, tt.cmd, logger)

			result, err := w.PushFiles()

			if !cmp.Equal(err, tt.err) {
				t.Errorf("err = %v, want %v", err, tt.err)
			}

			if !cmp.Equal(result, tt.want) {
				t.Errorf("wrapper.PushFiles() = %v, want %v", result, tt.want)
			}
		})
	}
}
