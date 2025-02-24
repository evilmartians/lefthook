package lefthook

import (
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

type gitCmd struct{}

func (g gitCmd) WithEnv(string, string) system.Command {
	return g
}

func (g gitCmd) Run([]string, string, io.Reader, io.Writer, io.Writer) error {
	return nil
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
		name, hook, config     string
		gitArgs                []string
		envs                   map[string]string
		existingDirs           []string
		hookNameCompletions    []string
		hookCommandCompletions []string
		error                  bool
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
			hookNameCompletions: []string{"pre-commit"},
			error:               true,
		},
		{
			name: "Valid hook",
			hook: "pre-commit",
			config: `
pre-commit:
  parallel: false
  piped: true
`,
			hookNameCompletions: []string{"pre-commit"},
			error:               false,
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
			hookNameCompletions:    []string{"pre-commit"},
			hookCommandCompletions: []string{"echo"},
			error:                  false,
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
			hookNameCompletions:    []string{"pre-commit"},
			hookCommandCompletions: []string{"echo"},
			error:                  false,
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
			hookNameCompletions:    []string{"post-commit"},
			hookCommandCompletions: []string{"echo"},
			error:                  true,
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			assert := assert.New(t)
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{
				Options: &Options{Fs: fs},
				repo: &git.Repository{
					Fs:        fs,
					Git:       git.NewExecutor(gitCmd{}),
					HooksPath: hooksPath,
					RootPath:  root,
					GitPath:   gitPath,
				},
			}
			lefthook.repo.Setup()

			// Create files that should exist
			for _, path := range tt.existingDirs {
				assert.NoError(fs.MkdirAll(path, 0o755))
			}

			assert.NoError(afero.WriteFile(fs, configPath, []byte(tt.config), 0o644))
			for env, value := range tt.envs {
				t.Setenv(env, value)
			}

			git.ResetState()
			err = lefthook.Run(tt.hook, RunArgs{}, tt.gitArgs)
			if tt.error {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			hookNameCompletions := lefthook.configHookCompletions()
			assert.ElementsMatch(tt.hookNameCompletions, hookNameCompletions)

			hookCommandCompletions := lefthook.configHookCommandCompletions(tt.hook)
			assert.ElementsMatch(tt.hookCommandCompletions, hookCommandCompletions)
		})
	}
}
