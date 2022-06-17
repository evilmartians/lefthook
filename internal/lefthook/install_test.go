package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
)

func TestLefthookInstall(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	configPath := filepath.Join(root, "lefthook.yml")
	hooksPath := filepath.Join(root, ".git", "hooks")

	hookPath := func(hook string) string {
		return filepath.Join(root, ".git", "hooks", hook)
	}

	repo := &git.Repository{
		HooksPath: hooksPath,
		RootPath:  root,
	}

	for n, tt := range [...]struct {
		name, config            string
		args                    InstallArgs
		existingHooks           map[string]string
		wantExist, wantNotExist []string
		wantError               bool
	}{
		{
			name:      "without a config file",
			wantExist: []string{configPath},
		},
		{
			name: "simple default config",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.ChecksumHookName),
			},
		},
		{
			name: "with existing hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"pre-commit": "",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("pre-commit.old"),
				hookPath("post-commit"),
				hookPath(config.ChecksumHookName),
			},
		},
		{
			name: "with existing lefthook hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"pre-commit": "# LEFTHOOK file",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.ChecksumHookName),
			},
			wantNotExist: []string{
				hookPath("pre-commit.old"),
			},
		},
		{
			name: "with synchronized hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"prepare-commit-msg": "# lefthook_version: 8b2c9fc6b3391b3cf020b97ab7037c61",
			},
			wantExist: []string{
				configPath,
				hookPath(config.ChecksumHookName),
			},
			wantNotExist: []string{
				hookPath("pre-commit"),
				hookPath("post-commit"),
			},
		},
		{
			name: "with synchronized hooks forced",
			args: InstallArgs{Force: true},
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"prepare-commit-msg": "# lefthook_version: 8b2c9fc6b3391b3cf020b97ab7037c61",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.ChecksumHookName),
			},
		},
		{
			name: "with existing hook and .old file",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"pre-commit":     "",
				"pre-commit.old": "",
			},
			wantError: true,
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("pre-commit.old"),
			},
			wantNotExist: []string{
				hookPath(config.ChecksumHookName),
			},
		},
		{
			name: "with existing hook and .old file, but forced",
			args: InstallArgs{Force: true},
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingHooks: map[string]string{
				"pre-commit":     "",
				"pre-commit.old": "",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("pre-commit.old"),
				hookPath("post-commit"),
				hookPath(config.ChecksumHookName),
			},
		},
	} {
		fs := afero.NewMemMapFs()
		lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			// Create configuration file
			if len(tt.config) > 0 {
				if err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Create files that should exist
			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err := fs.MkdirAll(filepath.Base(path), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, path, []byte(content), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Do install
			err := lefthook.Install(&tt.args)
			if tt.wantError && err == nil {
				t.Errorf("expected an error")
			} else if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// Test files that should exist
			for _, file := range tt.wantExist {
				ok, err := afero.Exists(fs, file)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !ok {
					t.Errorf("expected %s to exist", file)
				}
			}

			// Test files that should not exist
			for _, file := range tt.wantNotExist {
				ok, err := afero.Exists(fs, file)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if ok {
					t.Errorf("expected %s to not exist", file)
				}
			}
		})
	}
}
