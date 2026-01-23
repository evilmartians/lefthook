package command

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

func TestLefthookInstall(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	configPath := filepath.Join(root, "lefthook.yml")

	hookPath := func(hook string) string {
		return filepath.Join(gittest.GitPath(root), "hooks", hook)
	}

	infoPath := func(file string) string {
		return filepath.Join(gittest.GitPath(root), "info", file)
	}

	for n, tt := range [...]struct {
		name, config, checksum  string
		force                   bool
		hooks                   []string
		git                     []cmdtest.Out
		existingFiles           map[string]string
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
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath(config.GhostHookName),
			},
		},
		{
			name: "with given hook",
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
			hooks: []string{"pre-commit"},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("post-commit"),
				hookPath(config.GhostHookName),
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
			existingFiles: map[string]string{
				hookPath("pre-commit"): "",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("pre-commit.old"),
				hookPath("post-commit"),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath(config.GhostHookName),
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
			existingFiles: map[string]string{
				hookPath("pre-commit"): "# LEFTHOOK file",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("pre-commit.old"),
				hookPath(config.GhostHookName),
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
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath(config.GhostHookName),
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
			existingFiles: map[string]string{
				hookPath("pre-commit"):     "",
				hookPath("pre-commit.old"): "",
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
			existingFiles: map[string]string{
				hookPath("pre-commit"):     "",
				hookPath("pre-commit.old"): "",
			},
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("pre-commit.old"),
				hookPath("post-commit"),
				infoPath(config.ChecksumFileName),
			},
		},
		{
			name: "with custom hook",
			config: `
my-custom-hook:
  commands:
    custom:
      run: echo 'Hello from custom!'
`,
			wantNotExist: []string{
				hookPath("my-custom-hook"),
			},
		},
		{
			name: "with custom existing hook",
			config: `
my-custom-hook:
  commands:
    custom:
      run: echo 'Hello from custom!'
`,
			existingFiles: map[string]string{
				hookPath("my-custom-hook"): "",
			},
			wantExist: []string{
				hookPath("my-custom-hook.old"),
			},
			wantNotExist: []string{
				hookPath("my-custom-hook"),
			},
		},
		{
			name: "with unfetched remote",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    configs:
      - lefthook.yml
`,
			git: []cmdtest.Out{
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes") + " clone --quiet --origin origin --depth 1 https://github.com/evilmartians/lefthook lefthook",
				},
			},
		},
		{
			name: "needs refetching",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    ref: v2.0.0
    configs:
      - lefthook.yml
`,
			existingFiles: map[string]string{
				filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.1", ".git", "FETCH_HEAD"): "",
			},
			git: []cmdtest.Out{
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes") + " clone --quiet --origin origin --depth 1 --branch v2.0.0 https://github.com/evilmartians/lefthook lefthook-v2.0.0",
				},
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.0") + " fetch --quiet --depth 1 origin v2.0.0",
				},
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.0") + " checkout FETCH_HEAD",
				},
			},
		},
	} {
		fs := afero.NewMemMapFs()

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			assert := assert.New(t)

			// Prepend git config commands required by getHooksPathConfig() in install.go.
			// These commands are always called at the start of Install() to detect core.hooksPath conflicts.
			gitCmds := tt.git
			if len(gitCmds) == 0 || gitCmds[0].Command != "git config --local core.hooksPath" {
				gitCmds = append([]cmdtest.Out{
					{Command: "git config --local core.hooksPath"},
					{Command: "git config --global core.hooksPath"},
				}, gitCmds...)
			}

			repo := gittest.NewRepositoryBuilder().
				Root(root).
				Fs(fs).
				Cmd(cmdtest.NewOrdered(t, gitCmds)).
				Build()
			lefthook := &Lefthook{
				fs:   fs,
				repo: repo,
			}

			// Create configuration file
			if len(tt.config) > 0 {
				assert.NoError(afero.WriteFile(fs, configPath, []byte(tt.config), 0o644))
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				assert.NoError(fs.Chtimes(configPath, timestamp, timestamp))
			}

			if len(tt.checksum) > 0 {
				assert.NoError(afero.WriteFile(fs, lefthook.checksumFilePath(), []byte(tt.checksum), 0o644))
			}

			// Create files that should exist
			for path, content := range tt.existingFiles {
				assert.NoError(fs.MkdirAll(filepath.Dir(path), 0o755))
				assert.NoError(afero.WriteFile(fs, path, []byte(content), 0o755))
			}

			// Do install
			err := lefthook.Install(t.Context(), InstallArgs{Force: tt.force}, tt.hooks)
			if tt.wantError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			// Test files that should exist
			for _, file := range tt.wantExist {
				ok, err := afero.Exists(fs, file)
				assert.NoError(err)
				assert.Equal(true, ok)
			}

			// Test files that should not exist
			for _, file := range tt.wantNotExist {
				ok, err := afero.Exists(fs, file)
				assert.NoError(err)
				assert.Equal(false, ok)
			}
		})
	}
}

func Test_syncHooks(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	configPath := filepath.Join(root, "lefthook.yml")

	hookPath := func(hook string) string {
		return filepath.Join(root, ".git", "hooks", hook)
	}

	infoPath := func(file string) string {
		return filepath.Join(root, ".git", "info", file)
	}

	for n, tt := range [...]struct {
		name, config, checksum  string
		existingFiles           map[string]string
		git                     []cmdtest.Out
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
		{
			name: "unsynchronized",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'

commit-msg:
  jobs:
    - run: echo 'commit-msg'
`,
			checksum: "00000000f706df65f379a9ff5ce0119b 1555894311\n",
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				hookPath("commit-msg"),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath(config.GhostHookName),
			},
		},
		{
			name: "unsynchronized with selected hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'

commit-msg:
  jobs:
    - run: echo 'commit-msg'
`,
			checksum: "00000000f706df65f379a9ff5ce0119b 1555894310 pre-commit,post-commit\n",
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
				infoPath(config.ChecksumFileName),
			},
			wantNotExist: []string{
				hookPath("commit-msg"),
				hookPath(config.GhostHookName),
			},
		},
		{
			name: "with unfetched remote",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    configs:
      - lefthook.yml
`,
			git: []cmdtest.Out{
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes") + " clone --quiet --origin origin --depth 1 https://github.com/evilmartians/lefthook lefthook",
				},
			},
		},
		{
			name: "no need to refetch",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    ref: v2.0.1
    configs:
      - lefthook.yml
`,
			existingFiles: map[string]string{
				filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.1", ".git", "FETCH_HEAD"): "",
			},
		},
		{
			name: "needs refetching",
			config: `
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    ref: v2.0.0
    configs:
      - lefthook.yml
`,
			existingFiles: map[string]string{
				filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.1", ".git", "FETCH_HEAD"): "",
			},
			git: []cmdtest.Out{
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes") + " clone --quiet --origin origin --depth 1 --branch v2.0.0 https://github.com/evilmartians/lefthook lefthook-v2.0.0",
				},
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.0") + " fetch --quiet --depth 1 origin v2.0.0",
				},
				{
					Command: "git -C " + filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.0") + " checkout FETCH_HEAD",
				},
			},
			wantNotExist: []string{
				filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v2.0.1"),
			},
		},
	} {
		fs := afero.NewMemMapFs()

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			assert := assert.New(t)

			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Cmd(cmdtest.NewOrdered(t, tt.git)).Build()
			lefthook := &Lefthook{
				fs:   fs,
				repo: repo,
			}

			// Create configuration file
			if len(tt.config) > 0 {
				assert.NoError(afero.WriteFile(fs, configPath, []byte(tt.config), 0o644))
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				assert.NoError(fs.Chtimes(configPath, timestamp, timestamp))
			}

			if len(tt.checksum) > 0 {
				assert.NoError(afero.WriteFile(fs, lefthook.checksumFilePath(), []byte(tt.checksum), 0o644))
			}

			// Create files that should exist
			for path, content := range tt.existingFiles {
				assert.NoError(fs.MkdirAll(filepath.Dir(path), 0o755))
				assert.NoError(afero.WriteFile(fs, path, []byte(content), 0o755))
			}

			cfg, err := config.Load(lefthook.fs, repo)
			assert.NoError(err)

			// Create hooks
			_, err = lefthook.syncHooks(cfg, true)
			if tt.wantError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			// Test files that should exist
			for _, file := range tt.wantExist {
				ok, err := afero.Exists(fs, file)
				assert.NoError(err)
				assert.Equal(true, ok, file)
			}

			// Test files that should not exist
			for _, file := range tt.wantNotExist {
				ok, err := afero.Exists(fs, file)
				assert.NoError(err)
				assert.Equal(false, ok, file)
			}
		})
	}
}

func TestShouldRefetch(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	configPath := filepath.Join(root, "lefthook.yml")
	fetchHeadPath := func(lefthook *Lefthook, remote *config.Remote) string {
		remotePath := lefthook.repo.RemoteFolder(remote.GitURL, remote.Ref)
		return filepath.Join(remotePath, ".git", "FETCH_HEAD")
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
			shouldRefetchInitially: true,
			shouldRefetchAfter:     false,
			shouldRefetchBefore:    false,
		},
	} {
		fs := afero.NewMemMapFs()
		repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
		lefthook := &Lefthook{
			fs:   fs,
			repo: repo,
		}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			assert := assert.New(t)

			// Create configuration file
			if len(tt.config) > 0 {
				assert.NoError(afero.WriteFile(fs, configPath, []byte(tt.config), 0o644))
				timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
				assert.NoError(fs.Chtimes(configPath, timestamp, timestamp))
			}

			cfg, err := config.Load(lefthook.fs, repo)
			assert.NoError(err)

			remote := cfg.Remotes[0]

			assert.Equal(lefthook.shouldRefetch(remote), tt.shouldRefetchInitially)

			assert.NoError(afero.WriteFile(fs, fetchHeadPath(lefthook, remote), []byte(""), 0o644))
			firstFetchTime := time.Now().Add(-2 * time.Minute)

			assert.NoError(fs.Chtimes(fetchHeadPath(lefthook, remote), firstFetchTime, firstFetchTime))
			assert.Equal(lefthook.shouldRefetch(remote), tt.shouldRefetchAfter)

			assert.NoError(fs.Chtimes(fetchHeadPath(lefthook, remote), firstFetchTime, time.Now()))
			assert.Equal(lefthook.shouldRefetch(remote), tt.shouldRefetchBefore)
		})
	}
}

func TestLefthookInstallWithCoreHooksPath(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	configPath := filepath.Join(root, "lefthook.yml")

	hookPath := func(hook string) string {
		return filepath.Join(gittest.GitPath(root), "hooks", hook)
	}

	infoPath := func(file string) string {
		return filepath.Join(gittest.GitPath(root), "info", file)
	}

	configContent := `
pre-commit:
  commands:
    tests:
      run: yarn test
`

	for n, tt := range [...]struct {
		name         string
		force        bool
		git          []cmdtest.Out
		wantError    bool
		wantErrorMsg string
		wantExist    []string
	}{
		{
			name:  "with local and global core.hooksPath without --force",
			force: false,
			git: []cmdtest.Out{
				{
					Command: "git config --local core.hooksPath",
					Output:  ".custom-hooks",
				},
				{
					Command: "git config --global core.hooksPath",
					Output:  "/usr/local/hooks",
				},
			},
			wantError:    true,
			wantErrorMsg: "core.hooksPath",
		},
		{
			name:  "with local and global core.hooksPath with --force",
			force: true,
			git: []cmdtest.Out{
				{
					Command: "git config --local core.hooksPath",
					Output:  ".custom-hooks",
				},
				{
					Command: "git config --global core.hooksPath",
					Output:  "/usr/local/hooks",
				},
				{
					Command: "git config --local --unset-all core.hooksPath",
				},
				{
					Command: "git config --global --unset-all core.hooksPath",
				},
			},
			wantError: false,
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
				infoPath(config.ChecksumFileName),
			},
		},
	} {
		fs := afero.NewMemMapFs()

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			assert := assert.New(t)

			repo := gittest.NewRepositoryBuilder().
				Root(root).
				Fs(fs).
				Cmd(cmdtest.NewOrdered(t, tt.git)).
				Build()
			lefthook := &Lefthook{
				fs:   fs,
				repo: repo,
			}

			// Create configuration file
			assert.NoError(afero.WriteFile(fs, configPath, []byte(configContent), 0o644))
			timestamp := time.Date(2022, time.June, 22, 10, 40, 10, 1, time.UTC)
			assert.NoError(fs.Chtimes(configPath, timestamp, timestamp))

			// Do install
			err := lefthook.Install(t.Context(), InstallArgs{Force: tt.force}, nil)
			if tt.wantError {
				if assert.Error(err) && tt.wantErrorMsg != "" {
					assert.Contains(err.Error(), tt.wantErrorMsg)
				}
			} else {
				assert.NoError(err)
				// Test files that should exist
				for _, file := range tt.wantExist {
					ok, err := afero.Exists(fs, file)
					assert.NoError(err)
					assert.True(ok)
				}
			}
		})
	}
}
