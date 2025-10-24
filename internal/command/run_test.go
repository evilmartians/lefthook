package command

import (
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

type gitCmd struct{}

func (g gitCmd) WithoutEnvs(...string) system.Command {
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

	gitPath := gittest.GitPath(root)
	configPath := filepath.Join(root, "lefthook.yml")

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
				fs:   fs,
				repo: gittest.NewRepositoryBuilder().Git(gitCmd{}).Fs(fs).Root(root).Build(),
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

			err = lefthook.Run(t.Context(), RunArgs{Hook: tt.hook, GitArgs: tt.gitArgs})
			if tt.error {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			// hookNameCompletions := lefthook.configHookCompletions()
			// assert.ElementsMatch(tt.hookNameCompletions, hookNameCompletions)
			//
			// hookCommandCompletions := lefthook.configHookCommandCompletions(tt.hook)
			// assert.ElementsMatch(tt.hookCommandCompletions, hookCommandCompletions)
		})
	}
}
