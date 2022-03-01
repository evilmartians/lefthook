package lefthook

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestRun(t *testing.T) {
	for i, tt := range [...]struct {
		name     string
		global   []byte
		local    []byte
		scripts  map[string][]byte
		hookName string
		gitArgs  []string
		isErr    bool
		okList   []string
		failList []string
	}{
		{
			name: "scripts",
			global: []byte(`
source_dir: "/"
source_dir_local: "/"

pre-commit:
  commands:
    tests:
      runner: echo "test"
`),
			local: []byte(`
source_dir: "/"
source_dir_local: "/"

post-commit:
  commands:
    ping-done:
      run: echo "test"
`),
			scripts:  nil,
			hookName: "pre-commit",
			gitArgs:  nil,
			isErr:    false,
			okList:   []string{"tests"},
			failList: nil,
		},
		{
			name: "executable",
			global: []byte(`
source_dir: "/"
source_dir_local: "/local"

post-commit:
  scripts:
    "test.sh":
      runner: echo
`),
			local: nil,
			scripts: map[string][]byte{
				"test.sh": []byte(`echo "test"`),
			},
			hookName: "post-commit",
			gitArgs:  nil,
			isErr:    false,
			okList:   []string{"test.sh"},
			failList: nil,
		},

		{
			name: "parallel",
			global: []byte(`
post-commit:
  parallel: true
  commands:
    test0:
      run: echo one
    test1:
      run: echo two
    test2:
      run: ech three
`),
			local: nil,
			scripts: map[string][]byte{
				"test.sh": []byte(`echo "test"`),
			},
			hookName: "post-commit",
			gitArgs:  nil,
			isErr:    true,
			okList:   []string{"test0", "test1"},
			failList: []string{"test2"},
		},
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			defer dropState()

			if err := fs.WriteFile("/lefthook.yml", tt.global, 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if err := fs.WriteFile("/lefthook-local.yml", tt.local, 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for script, blob := range tt.scripts {
				if err := fs.WriteFile(fmt.Sprintf("/%s/%s", tt.hookName, script), blob, 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			repo, err := git.NewRepository()
			repo.RootPath = "/"
			repo.HooksPath = "/"
			require.NoError(t, err)
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			err = lefthook.Run(tt.hookName, tt.gitArgs)
			if tt.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tt.okList, okList)
			require.ElementsMatch(t, tt.failList, failList)
		})
	}
}

func Test_RunBrokenPipe(t *testing.T) {
	global := []byte(`
post-commit:
  piped: true
  commands:
    test0:
      run: non-existing-command
    test1:
      run: echo
`)
	defer dropState()
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	if err := fs.WriteFile("/lefthook.yml", global, 0o644); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	repo, err := git.NewRepository()
	repo.RootPath = "/"
	repo.HooksPath = "/"
	require.NoError(t, err)
	lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

	err = lefthook.Run("post-commit", nil)
	require.Error(t, err)

	require.ElementsMatch(t, nil, okList)
	require.ElementsMatch(t, []string{"test0"}, failList)
	require.Equal(t, true, isPipeBroken)
}

func dropState() {
	okList = nil
	failList = nil
	isPipeBroken = false
}
