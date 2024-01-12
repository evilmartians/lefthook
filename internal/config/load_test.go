package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

//gocyclo:ignore
func TestLoad(t *testing.T) {
	root, err := filepath.Abs("")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for i, tt := range [...]struct {
		name                  string
		global, local, remote string
		remoteConfigPath      string
		otherFiles            map[string]string
		result                *Config
	}{
		{
			name: "with global, dot",
			otherFiles: map[string]string{
				".lefthook.yml": `
pre-commit:
  commands:
    tests:
      run: yarn test
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "yarn test",
							},
						},
					},
				},
			},
		},
		{
			name: "with global, nodot",
			otherFiles: map[string]string{
				"lefthook.yml": `
pre-commit:
  commands:
    tests:
      run: yarn test
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "yarn test",
							},
						},
					},
				},
			},
		},
		{
			name: "with global, nodot has priority",
			otherFiles: map[string]string{
				".lefthook.yml": `
pre-commit:
  commands:
    tests:
      run: yarn test1
`,
				"lefthook.yml": `
pre-commit:
  commands:
    tests:
      run: yarn test2
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "yarn test2",
							},
						},
					},
				},
			},
		},
		{
			name: "simple",
			global: `
pre-commit:
  commands:
    tests:
      run: yarn test
`,
			local: `
post-commit:
  commands:
    ping-done:
      run: curl -x POST status.com/done
`,
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "yarn test",
							},
						},
					},
					"post-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"ping-done": {
								Run: "curl -x POST status.com/done",
							},
						},
					},
				},
			},
		},
		{
			name: "with overrides",
			global: `
min_version: 0.6.0
source_dir: $HOME/sources
source_dir_local: $HOME/sources_local

pre-commit:
  parallel: true
  commands:
    tests:
      run: bundle exec rspec
      tags: [backend, test]
    lint:
      run: bundle exec rubocop
      glob: "*.rb"
      tags: [backend, linter]
  scripts:
    "format.sh":
      runner: bash
`,
			local: `
min_version: 1.0.0
colors: false

pre-commit:
  commands:
    tests:
      skip: true
    lint:
      run: docker exec -it ruby:2.7 {cmd}
  scripts:
    "format.sh":
      only: true

pre-push:
  commands:
    rubocop:
      run: bundle exec rubocop
      tags: [backend, linter]
`,
			result: &Config{
				MinVersion:     "1.0.0",
				Colors:         false,
				SourceDir:      "$HOME/sources",
				SourceDirLocal: "$HOME/sources_local",

				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: true,
						Commands: map[string]*Command{
							"tests": {
								Skip: true,
								Run:  "bundle exec rspec",
								Tags: []string{"backend", "test"},
							},
							"lint": {
								Glob: "*.rb",
								Run:  "docker exec -it ruby:2.7 bundle exec rubocop",
								Tags: []string{"backend", "linter"},
							},
						},
						Scripts: map[string]*Script{
							"format.sh": {
								Only:   true,
								Runner: "bash",
							},
						},
					},
					"pre-push": {
						Commands: map[string]*Command{
							"rubocop": {
								Run:  "bundle exec rubocop",
								Tags: []string{"backend", "linter"},
							},
						},
					},
				},
			},
		},
		{
			name: "with overrides, dot",
			otherFiles: map[string]string{
				".lefthook.yml": `
pre-push:
  scripts:
    "global-extend":
      runner: bash
`,
				".lefthook-local.yml": `
pre-push:
  scripts:
    "local-extend":
      runner: bash
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-push": {
						Scripts: map[string]*Script{
							"global-extend": {
								Runner: "bash",
							},
							"local-extend": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		{
			name: "with overrides, dot, nodot",
			otherFiles: map[string]string{
				"lefthook.yml": `
pre-push:
  scripts:
    "global-extend":
      runner: bash
`,
				".lefthook-local.yml": `
pre-push:
  scripts:
    "local-extend":
      runner: bash
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-push": {
						Scripts: map[string]*Script{
							"global-extend": {
								Runner: "bash",
							},
							"local-extend": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		{
			name: "with overrides, nodot has priority",
			otherFiles: map[string]string{
				"lefthook.yml": `
pre-push:
  scripts:
    "global-extend":
      runner: bash
`,
				".lefthook-local.yml": `
pre-push:
  scripts:
    "local-extend":
      runner: bash1
`,
				"lefthook-local.yml": `
pre-push:
  scripts:
    "local-extend":
      runner: bash2
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-push": {
						Scripts: map[string]*Script{
							"global-extend": {
								Runner: "bash",
							},
							"local-extend": {
								Runner: "bash2",
							},
						},
					},
				},
			},
		},
		{
			name: "with extra hooks",
			global: `
tests:
  commands:
    tests:
      run: go test ./...

lints:
  scripts:
    "linter.sh":
      runner: bash
`,
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Hooks: map[string]*Hook{
					"tests": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "go test ./...",
							},
						},
					},
					"lints": {
						Scripts: map[string]*Script{
							"linter.sh": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		{
			name: "with extra hooks only in local config",
			global: `
colors:
  yellow: '#FFE4B5'
  red: 196
tests:
  commands:
    tests:
      run: go test ./...
`,
			local: `
lints:
  scripts:
    "linter.sh":
      runner: bash
`,
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         map[string]interface{}{"yellow": "#FFE4B5", "red": 196},
				Hooks: map[string]*Hook{
					"tests": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run: "go test ./...",
							},
						},
					},
					"lints": {
						Scripts: map[string]*Script{
							"linter.sh": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		{
			name: "with remote",
			global: `
remote:
  git_url: git@github.com:evilmartians/lefthook
`,
			remote: `
pre-commit:
  commands:
    lint:
      run: yarn lint
  scripts:
    "test.sh":
      runner: bash
`,
			remoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook", "lefthook.yml"),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Remote: &Remote{
					GitURL: "git@github.com:evilmartians/lefthook",
				},
				Hooks: map[string]*Hook{
					"pre-commit": {
						Commands: map[string]*Command{
							"lint": {
								Run: "yarn lint",
							},
						},
						Scripts: map[string]*Script{
							"test.sh": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		{
			name: "with remote and custom config name",
			global: `
remote:
  git_url: git@github.com:evilmartians/lefthook
  ref: v1.0.0
  config: examples/custom.yml

pre-commit:
  only:
    - ref: main
  commands:
    global:
      run: echo 'Global!'
    lint:
      run: this will be overwritten
`,
			remote: `
pre-commit:
  commands:
    lint:
      only:
        - merge
        - rebase
      run: yarn lint
  scripts:
    "test.sh":
      skip:
        - merge
      runner: bash
`,
			remoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v1.0.0", "examples", "custom.yml"),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Remote: &Remote{
					GitURL: "git@github.com:evilmartians/lefthook",
					Ref:    "v1.0.0",
					Config: "examples/custom.yml",
				},
				Hooks: map[string]*Hook{
					"pre-commit": {
						Only: []interface{}{map[string]interface{}{"ref": "main"}},
						Commands: map[string]*Command{
							"lint": {
								Run:  "yarn lint",
								Only: []interface{}{"merge", "rebase"},
							},
							"global": {
								Run: "echo 'Global!'",
							},
						},
						Scripts: map[string]*Script{
							"test.sh": {
								Runner: "bash",
								Skip:   []interface{}{"merge"},
							},
						},
					},
				},
			},
		},
		{
			name: "with extends",
			global: `
extends:
  - global-extend.yml

remote:
  git_url: https://github.com/evilmartians/lefthook
  config: examples/config.yml

pre-push:
  commands:
    global:
      run: echo global
`,
			local: `
extends:
  - local-extend.yml

pre-push:
  commands:
    local:
      run: echo local
`,
			remote: `
extends:
  - ../remote-extend.yml

pre-push:
  commands:
    remote:
      run: echo remote
`,
			remoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook", "examples", "config.yml"),
			otherFiles: map[string]string{
				"global-extend.yml": `
pre-push:
  scripts:
    "global-extend":
      runner: bash
`,
				"local-extend.yml": `
pre-push:
  scripts:
    "local-extend":
      runner: bash
`,
				".git/info/lefthook-remotes/lefthook/remote-extend.yml": `
pre-push:
  scripts:
    "remote-extend":
      runner: bash
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Remote: &Remote{
					GitURL: "https://github.com/evilmartians/lefthook",
					Config: "examples/config.yml",
				},
				Extends: []string{"local-extend.yml"},
				Hooks: map[string]*Hook{
					"pre-push": {
						Commands: map[string]*Command{
							"global": {
								Run: "echo global",
							},
							"local": {
								Run: "echo local",
							},
							"remote": {
								Run: "echo remote",
							},
						},
						Scripts: map[string]*Script{
							"global-extend": {
								Runner: "bash",
							},
							"local-extend": {
								Runner: "bash",
							},
							"remote-extend": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		repo := &git.Repository{
			Fs:       fs,
			RootPath: root,
			InfoPath: filepath.Join(root, ".git", "info"),
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			if tt.global != "" {
				if err := fs.WriteFile(filepath.Join(root, "lefthook.yml"), []byte(tt.global), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if tt.local != "" {
				if err := fs.WriteFile(filepath.Join(root, "lefthook-local.yml"), []byte(tt.local), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if len(tt.remoteConfigPath) > 0 {
				if err := fs.MkdirAll(filepath.Base(tt.remoteConfigPath), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if err := fs.WriteFile(tt.remoteConfigPath, []byte(tt.remote), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			for name, content := range tt.otherFiles {
				path := filepath.Join(
					root,
					filepath.Join(strings.Split(name, "/")...),
				)
				dir := filepath.Dir(path)

				if err := fs.MkdirAll(dir, 0o775); err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if err := fs.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			checkConfig, err := Load(fs.Fs, repo)

			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
			}
		})
	}

	for i, tt := range [...]struct {
		name             string
		yaml, json, toml string
		result           *Config
	}{
		{
			name: "simple configs",
			yaml: `
pre-commit:
  commands:
    echo:
      run: echo 1
`,
			json: `
{
  "pre-commit": {
    "commands": {
      "echo": { "run": "echo 1" }
    }
  }
}`,
			toml: `
[pre-commit.commands.echo]
run = "echo 1"
`,
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Commands: map[string]*Command{
							"echo": {
								Run: "echo 1",
							},
						},
					},
				},
			},
		},
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		repo := &git.Repository{
			Fs:       fs,
			RootPath: root,
			InfoPath: filepath.Join(root, ".git", "info"),
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			yamlConfig := filepath.Join(root, "lefthook.yml")
			if err := fs.WriteFile(yamlConfig, []byte(tt.yaml), 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			checkConfig, err := Load(fs.Fs, repo)
			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
			}

			if err = fs.Remove(yamlConfig); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			jsonConfig := filepath.Join(root, "lefthook.json")
			if err = fs.WriteFile(jsonConfig, []byte(tt.json), 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			checkConfig, err = Load(fs.Fs, repo)
			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
			}

			if err = fs.Remove(jsonConfig); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			tomlConfig := filepath.Join(root, "lefthook.toml")
			if err = fs.WriteFile(tomlConfig, []byte(tt.toml), 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			checkConfig, err = Load(fs.Fs, repo)
			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
			}

			if err = fs.Remove(tomlConfig); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}

	type remote struct {
		RemoteConfigPath string
		Content          string
	}
	for i, tt := range [...]struct {
		name          string
		global, local string
		remotes       []remote
		otherFiles    map[string]string
		result        *Config
	}{
		{
			name: "with remotes, config and configs",
			global: `
pre-commit:
  only:
    - ref: main
  commands:
    global:
      run: echo 'Global!'
    lint:
      run: this will be overwritten
remotes:
  - git_url: https://github.com/evilmartians/lefthook
    ref: v1.0.0
    config: examples/custom.yml
  - git_url: https://github.com/evilmartians/lefthook
    configs:
      - examples/remote/ping.yml
    ref: v1.5.5
`,
			remotes: []remote{
				{
					RemoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v1.0.0", "examples", "custom.yml"),
					Content: `
pre-commit:
  commands:
    lint:
      only:
        - merge
        - rebase
      run: yarn lint
  scripts:
    "test.sh":
      skip:
        - merge
      runner: bash
`,
				},
				{
					RemoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook-v1.5.5", "examples", "remote", "ping.yml"),
					Content: `
pre-commit:
  commands:
    ping:
      run: echo pong
`,
				},
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Remotes: []*Remote{
					{
						GitURL: "https://github.com/evilmartians/lefthook",
						Ref:    "v1.0.0",
						Config: "examples/custom.yml",
					},
					{
						GitURL: "https://github.com/evilmartians/lefthook",
						Ref:    "v1.5.5",
						Configs: []string{
							"examples/remote/ping.yml",
						},
					},
				},
				Hooks: map[string]*Hook{
					"pre-commit": {
						Only: []interface{}{map[string]interface{}{"ref": "main"}},
						Commands: map[string]*Command{
							"lint": {
								Run:  "yarn lint",
								Only: []interface{}{"merge", "rebase"},
							},
							"ping": {
								Run: "echo pong",
							},
							"global": {
								Run: "echo 'Global!'",
							},
						},
						Scripts: map[string]*Script{
							"test.sh": {
								Runner: "bash",
								Skip:   []interface{}{"merge"},
							},
						},
					},
				},
			},
		},
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		repo := &git.Repository{
			Fs:       fs,
			RootPath: root,
			InfoPath: filepath.Join(root, ".git", "info"),
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			if tt.global != "" {
				if err := fs.WriteFile(filepath.Join(root, "lefthook.yml"), []byte(tt.global), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if tt.local != "" {
				if err := fs.WriteFile(filepath.Join(root, "lefthook-local.yml"), []byte(tt.local), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			for _, remote := range tt.remotes {
				if err := fs.MkdirAll(filepath.Base(remote.RemoteConfigPath), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if err := fs.WriteFile(remote.RemoteConfigPath, []byte(remote.Content), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			for name, content := range tt.otherFiles {
				path := filepath.Join(
					root,
					filepath.Join(strings.Split(name, "/")...),
				)
				dir := filepath.Dir(path)

				if err := fs.MkdirAll(dir, 0o775); err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if err := fs.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			checkConfig, err := Load(fs.Fs, repo)

			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
			}
		})
	}
}
