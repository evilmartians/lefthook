package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

type GitMock struct{}

func (g GitMock) SetRootPath(_root string) {}

func (g GitMock) Cmd(_cmd string) (string, error) {
	return "", nil
}

func (g GitMock) CmdArgs(_args ...string) (string, error) {
	return "", nil
}

func (g GitMock) CmdLines(_cmd string) ([]string, error) {
	return nil, nil
}

func (g GitMock) RawCmd(_cmd string) (string, error) {
	return "", nil
}

func TestRun(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	configPath := filepath.Join(root, "lefthook.yml")
	hooksPath := filepath.Join(root, ".git", "hooks")
	gitPath := filepath.Join(root, ".git")

	for i, tt := range [...]struct {
		name, hook, config string
		gitArgs            []string
		envs               map[string]string
		existingDirs       []string
		error              bool
	}{
		{
			name: "Skip case",
			hook: "pre-commit",
			envs: map[string]string{
				"LEFTHOOK": "0",
			},
			error: false,
		},
		{
			name: "Skip case",
			hook: "pre-commit",
			envs: map[string]string{
				"LEFTHOOK": "false",
			},
			error: false,
		},
		{
			name: "Invalid version",
			hook: "pre-commit",
			config: `
min_version: 23.0.1
`,
			error: true,
		},
		{
			name: "Valid version, no hook",
			hook: "pre-commit",
			config: `
min_version: 0.7.9
`,
			error: false,
		},
		{
			name: "Invalid hook",
			hook: "pre-commit",
			config: `
pre-commit:
  parallel: true
  piped: true
`,
			error: true,
		},
		{
			name: "Valid hook",
			hook: "pre-commit",
			config: `
pre-commit:
  parallel: false
  piped: true
`,
			error: false,
		},
		{
			name: "When in git rebase-merge flow",
			hook: "pre-commit",
			config: `
pre-commit:
  parallel: false
  piped: true
  commands:
    echo:
      skip:
        - rebase
        - merge
      run: echo 'SHOULD NEVER RUN'
`,
			existingDirs: []string{
				filepath.Join(gitPath, "rebase-merge"),
			},
			error: false,
		},
		{
			name: "When in git rebase-apply flow",
			hook: "pre-commit",
			config: `
pre-commit:
  parallel: false
  piped: true
  commands:
    echo:
      skip:
        - rebase
        - merge
      run: echo 'SHOULD NEVER RUN'
`,
			existingDirs: []string{
				filepath.Join(gitPath, "rebase-apply"),
			},
			error: false,
		},
		{
			name: "When not in rebase flow",
			hook: "post-commit",
			config: `
post-commit:
  parallel: false
  piped: true
  commands:
    echo:
      skip:
        - rebase
        - merge
      run: echo 'SHOULD RUN'
`,
			error: true,
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{
				Options: &Options{Fs: fs},
				repo: &git.Repository{
					Fs:        fs,
					Git:       GitMock{},
					HooksPath: hooksPath,
					RootPath:  root,
					GitPath:   gitPath,
				},
			}

			// Create files that should exist
			for _, path := range tt.existingDirs {
				if err := fs.MkdirAll(path, 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for env, value := range tt.envs {
				t.Setenv(env, value)
			}

			err = lefthook.Run(tt.hook, RunArgs{}, tt.gitArgs)
			if err != nil {
				if !tt.error {
					t.Errorf("unexpected error: %s", err)
				}
			} else {
				if tt.error {
					t.Errorf("expected an error")
				}
			}
		})
	}
}
