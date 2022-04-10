package lefthook

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

// Common vars and functions for tests

const root = string(os.PathSeparator) + "src"

var (
	configPath = filepath.Join(root, "lefthook.yml")
	hooksPath  = filepath.Join(root, ".git", "hooks")
)

func hookPath(hook string) string {
	return filepath.Join(root, ".git", "hooks", hook)
}

//

func TestRun(t *testing.T) {
	repo := &git.Repository{
		HooksPath: hooksPath,
		RootPath:  root,
	}

	for i, tt := range [...]struct {
		name, hook, config string
		gitArgs            []string
		envs               map[string]string
		error              bool
	}{
		{
			name: "Skip case",
			hook: "any-hook",
			envs: map[string]string{
				"LEFTHOOK": "0",
			},
			error: false,
		},
		{
			name: "Skip case",
			hook: "any-hook",
			envs: map[string]string{
				"LEFTHOOK": "false",
			},
			error: false,
		},
		{
			name: "Invalid version",
			hook: "any-hook",
			config: `
min_version: 23.0.1
`,
			error: true,
		},
		{
			name: "Valid version, no hook",
			hook: "any-hook",
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
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for env, value := range tt.envs {
				t.Setenv(env, value)
			}

			err = lefthook.Run(tt.hook, tt.gitArgs)
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
