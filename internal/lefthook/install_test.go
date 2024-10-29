package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

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

	hookPath := func(hook string) string {
		return filepath.Join(root, ".git", "hooks", hook)
	}

	infoPath := func(file string) string {
		return filepath.Join(root, ".git", "info", file)
	}

	repo := &git.Repository{
		HooksPath: filepath.Join(root, ".git", "hooks"),
		RootPath:  root,
		InfoPath:  filepath.Join(root, ".git", "info"),
	}

	for n, tt := range [...]struct {
		name, config, checksum  string
		force                   bool
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
				hookPath(config.GhostHookName),
				infoPath(config.ChecksumFileName),
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
				hookPath(config.GhostHookName),
				infoPath(config.ChecksumFileName),
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
				hookPath(config.GhostHookName),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("pre-commit.old"),
			},
		},
		{
			name: "with stale timestamp and checksum",
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
			checksum: "8b2c9fc6b3391b3cf020b97ab7037c62 1555894310\n",
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.GhostHookName),
				infoPath(config.ChecksumFileName),
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
				infoPath(config.ChecksumFileName),
			},
		},
		{
			name:  "with existing hook and .old file, but forced",
			force: true,
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
				infoPath(config.ChecksumFileName),
			},
		},
	} {
		fs := afero.NewMemMapFs()
		lefthook := &Lefthook{
			Options: &Options{Fs: fs},
			repo:    repo,
		}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			// Create configuration file
			if len(tt.config) > 0 {
				if err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				if err := fs.Chtimes(configPath, timestamp, timestamp); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if len(tt.checksum) > 0 {
				if err := afero.WriteFile(fs, lefthook.checksumFilePath(), []byte(tt.checksum), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Create files that should exist
			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err := fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, path, []byte(content), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Do install
			err := lefthook.Install(tt.force)
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

func TestCreateHooksIfNeeded(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	configPath := filepath.Join(root, "lefthook.yml")

	hookPath := func(hook string) string {
		return filepath.Join(root, ".git", "hooks", hook)
	}

	infoPath := func(file string) string {
		return filepath.Join(root, ".git", "info", file)
	}

	repo := &git.Repository{
		HooksPath: filepath.Join(root, ".git", "hooks"),
		RootPath:  root,
		InfoPath:  filepath.Join(root, ".git", "info"),
	}
	for n, tt := range [...]struct {
		name, config, checksum  string
		existingHooks           map[string]string
		wantExist, wantNotExist []string
		wantError               bool
	}{
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
			checksum: "8b2c9fc6b3391b3cf020b97ab7037c61 1655894410\n",
			wantExist: []string{
				configPath,
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.GhostHookName),
			},
		},
		{
			name: "with stale timestamp but synchronized",
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
			checksum: "939f59e3f706df65f379a9ff5ce0119b 1555894310\n",
			wantExist: []string{
				configPath,
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath(config.GhostHookName),
			},
		},
	} {
		fs := afero.NewMemMapFs()
		lefthook := &Lefthook{
			Options: &Options{Fs: fs},
			repo:    repo,
		}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			// Create configuration file
			if len(tt.config) > 0 {
				if err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				if err := fs.Chtimes(configPath, timestamp, timestamp); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if len(tt.checksum) > 0 {
				if err := afero.WriteFile(fs, lefthook.checksumFilePath(), []byte(tt.checksum), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Create files that should exist
			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err := fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, path, []byte(content), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			cfg, err := config.Load(lefthook.Fs, repo)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// Create hooks
			err = lefthook.createHooksIfNeeded(cfg, true, true)
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

func TestShouldRefetch(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	configPath := filepath.Join(root, "lefthook.yml")
	fetchHeadPath := func(lefthook *Lefthook, remote *config.Remote) string {
		remotePath := lefthook.repo.RemoteFolder(remote.GitURL, remote.Ref)
		return filepath.Join(remotePath, ".git", "FETCH_HEAD")
	}

	repo := &git.Repository{
		HooksPath: filepath.Join(root, ".git", "hooks"),
		RootPath:  root,
		InfoPath:  filepath.Join(root, ".git", "info"),
	}
	for n, tt := range [...]struct {
		name, config                                                    string
		shouldRefetchInitially, shouldRefetchAfter, shouldRefetchBefore bool
	}{
		{
			name: "with refetch frequency configured to always",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch_frequency: always
    configs:
      - examples/remote/ping.yml
`,
			shouldRefetchInitially: true,
			shouldRefetchAfter:     true,
			shouldRefetchBefore:    true,
		},
		{
			name: "with refetch frequency configured to 1 minute",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch_frequency: 1m
    configs:
      - examples/remote/ping.yml
`,
			shouldRefetchInitially: true,
			shouldRefetchAfter:     true,
			shouldRefetchBefore:    false,
		},
		{
			name: "with refetch frequency configured to never",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch_frequency: never
    configs:
      - examples/remote/ping.yml
`,
			shouldRefetchInitially: false,
			shouldRefetchAfter:     false,
			shouldRefetchBefore:    false,
		},
		{
			name: "with refetch frequency not configured",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    configs:
      - examples/remote/ping.yml
`,
			shouldRefetchInitially: false,
			shouldRefetchAfter:     false,
			shouldRefetchBefore:    false,
		},
	} {
		fs := afero.NewMemMapFs()
		lefthook := &Lefthook{
			Options: &Options{Fs: fs},
			repo:    repo,
		}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			// Create configuration file
			if len(tt.config) > 0 {
				if err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				if err := fs.Chtimes(configPath, timestamp, timestamp); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			cfg, err := config.Load(lefthook.Fs, repo)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			remote := cfg.Remotes[0]

			if lefthook.shouldRefetch(remote) != tt.shouldRefetchInitially {
				t.Errorf("unexpected shouldRefetch return before first fetch")
			}

			if err := afero.WriteFile(fs, fetchHeadPath(lefthook, remote), []byte(""), 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			firstFetchTime := time.Now().Add(-2 * time.Duration(time.Minute))

			if err := fs.Chtimes(fetchHeadPath(lefthook, remote), firstFetchTime, firstFetchTime); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if lefthook.shouldRefetch(remote) != tt.shouldRefetchAfter {
				t.Errorf("unexpected shouldRefetch return after refetch period")
			}

			if err := fs.Chtimes(fetchHeadPath(lefthook, remote), firstFetchTime, time.Now()); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if lefthook.shouldRefetch(remote) != tt.shouldRefetchBefore {
				t.Errorf("unexpected shouldRefetch return before refetch period")
			}
		})
	}
}
