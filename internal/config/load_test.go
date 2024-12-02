package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/git"
)

//gocyclo:ignore
func TestLoad(t *testing.T) {
	root, err := filepath.Abs("")
	assert.NoError(t, err)

	for name, tt := range map[string]struct {
		files            map[string]string
		remote           string
		remoteConfigPath string
		result           *Config
	}{
		"with .lefthook.yml": {
			files: map[string]string{
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
		"with lefthook.yml": {
			files: map[string]string{
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
		"with lefthook.yml and .lefthook.yml": {
			files: map[string]string{
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
		"simple": {
			files: map[string]string{
				"lefthook.yml": `
pre-commit:
  commands:
    tests:
      run: yarn test
`,
				"lefthook-local.yml": `
post-commit:
  commands:
    ping-done:
      run: curl -x POST status.com/done
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
		"with overrides": {
			files: map[string]string{
				"lefthook.yml": `
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
				"lefthook-local.yml": `
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
			},
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
		"with overrides from .lefthook-local.yml": {
			files: map[string]string{
				".lefthook.yml": `
pre-push:
  scripts:
    "global-extend.sh":
      runner: bash
`,
				".lefthook-local.yml": `
pre-push:
  scripts:
    "local-extend.sh":
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
							"global-extend.sh": {
								Runner: "bash",
							},
							"local-extend.sh": {
								Runner: "bash",
							},
						},
					},
				},
			},
		},
		"with overrides, dot, nodot": {
			files: map[string]string{
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
		"with overrides, nodot has priority": {
			files: map[string]string{
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
		"with extra hooks": {
			files: map[string]string{
				"lefthook.yml": `
tests:
  commands:
    tests:
      run: go test ./...

lints:
  scripts:
    "linter.sh":
      runner: bash
`,
			},
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
		"with extra hooks only in local config": {
			files: map[string]string{
				"lefthook.yml": `
colors:
  yellow: '#FFE4B5'
  red: 196
tests:
  commands:
    tests:
      run: go test ./...
`,
				"lefthook-local.yml": `
lints:
  scripts:
    "linter.sh":
      runner: bash
`,
			},
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
		"with remote": {
			files: map[string]string{
				"lefthook.yml": `
remote:
  git_url: git@github.com:evilmartians/lefthook
`,
			},
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
				Remotes: []*Remote{
					{
						GitURL: "git@github.com:evilmartians/lefthook",
					},
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
		"with remote and custom config name": {
			files: map[string]string{
				"lefthook.yml": `
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
			},
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
				Remotes: []*Remote{
					{
						GitURL:  "git@github.com:evilmartians/lefthook",
						Ref:     "v1.0.0",
						Configs: []string{"examples/custom.yml"},
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
		"with extends": {
			files: map[string]string{
				"lefthook.yml": `
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
				"lefthook-local.yml": `
extends:
  - local-extend.yml

pre-push:
  commands:
    local:
      run: echo local
`,
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
			remote: `
extends:
  - ../remote-extend.yml

pre-push:
  commands:
    remote:
      run: echo remote
`,
			remoteConfigPath: filepath.Join(root, ".git", "info", "lefthook-remotes", "lefthook", "examples", "config.yml"),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         nil,
				Remotes: []*Remote{
					{
						GitURL:  "https://github.com/evilmartians/lefthook",
						Configs: []string{"examples/config.yml"},
					},
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
		"with extends and local": {
			files: map[string]string{
				"lefthook.yml": `
extends:
  - global-extend.yml
pre-commit:
  parallel: true
  exclude_tags: [linter]
  commands:
    global-lint:
      run: bundle exec rubocop
      glob: "*.rb"
      tags: [backend, linter]
    global-other:
      run: bundle exec rubocop
      tags: [other]
`,
				"lefthook-local.yml": `
pre-commit:
  exclude_tags: [backend]
`,
				"global-extend.yml": `
pre-commit:
  exclude_tags: [test]
  commands:
    extended-tests:
      run: bundle exec rspec
      tags: [backend, test]
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Extends:        []string{"global-extend.yml"},
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel:    true,
						ExcludeTags: []string{"backend"},
						Commands: map[string]*Command{
							"global-lint": {
								Run:  "bundle exec rubocop",
								Tags: []string{"backend", "linter"},
								Glob: "*.rb",
							},
							"global-other": {
								Run:  "bundle exec rubocop",
								Tags: []string{"other"},
							},
							"extended-tests": {
								Run:  "bundle exec rspec",
								Tags: []string{"backend", "test"},
							},
						},
					},
				},
			},
		},
		"with glob in extends": {
			files: map[string]string{
				"lefthook.yml": `
extends:
  - dir/*/config.yml
`,
				"dir/a/config.yml": `
pre-commit:
  commands:
    a:
      run: echo A
`,
				"dir/b/config.yml": `
pre-commit:
  commands:
    b:
      run: echo B
`,
				"dir/b/c/config.yml": `
pre-commit:
  commands:
    c:
      run: echo C
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Extends:        []string{"dir/*/config.yml"},
				Colors:         nil,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"a": {
								Run: "echo A",
							},
							"b": {
								Run: "echo B",
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

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			for name, content := range tt.files {
				path := filepath.Join(
					root,
					filepath.Join(strings.Split(name, "/")...),
				)
				dir := filepath.Dir(path)

				assert.NoError(fs.MkdirAll(dir, 0o775))
				assert.NoError(fs.WriteFile(path, []byte(content), 0o644))
			}

			if len(tt.remoteConfigPath) > 0 {
				assert.NoError(fs.MkdirAll(filepath.Base(tt.remoteConfigPath), 0o755))
				assert.NoError(fs.WriteFile(tt.remoteConfigPath, []byte(tt.remote), 0o644))
			}

			result, err := Load(fs.Fs, repo)
			assert.NoError(err)
			assert.Equal(result, tt.result)
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
			assert := assert.New(t)

			// YAML
			yamlConfig := filepath.Join(root, "lefthook.yml")
			assert.NoError(fs.WriteFile(yamlConfig, []byte(tt.yaml), 0o644))

			result, err := Load(fs.Fs, repo)
			assert.NoError(err)
			assert.Equal(result, tt.result)

			assert.NoError(fs.Remove(yamlConfig))

			// JSON
			jsonConfig := filepath.Join(root, "lefthook.json")
			assert.NoError(fs.WriteFile(jsonConfig, []byte(tt.json), 0o644))

			result, err = Load(fs.Fs, repo)
			assert.NoError(err)
			assert.Equal(result, tt.result)

			assert.NoError(fs.Remove(jsonConfig))

			// TOML
			tomlConfig := filepath.Join(root, "lefthook.toml")
			assert.NoError(fs.WriteFile(tomlConfig, []byte(tt.toml), 0o644))

			result, err = Load(fs.Fs, repo)
			assert.NoError(err)
			assert.Equal(result, tt.result)

			assert.NoError(fs.Remove(tomlConfig))
		})
	}

	type remote struct {
		RemoteConfigPath string
		Content          string
	}
	for name, tt := range map[string]struct {
		files   map[string]string
		remotes []remote
		result  *Config
	}{
		"with remotes, config and configs": {
			files: map[string]string{
				"lefthook.yml": `
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
			},
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
						GitURL:  "https://github.com/evilmartians/lefthook",
						Ref:     "v1.0.0",
						Configs: []string{"examples/custom.yml"},
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

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			for name, content := range tt.files {
				path := filepath.Join(
					root,
					filepath.Join(strings.Split(name, "/")...),
				)
				dir := filepath.Dir(path)

				assert.NoError(fs.MkdirAll(dir, 0o775))
				assert.NoError(fs.WriteFile(path, []byte(content), 0o644))
			}

			for _, remote := range tt.remotes {
				assert.NoError(fs.MkdirAll(filepath.Base(remote.RemoteConfigPath), 0o755))
				assert.NoError(fs.WriteFile(remote.RemoteConfigPath, []byte(remote.Content), 0o644))
			}

			result, err := Load(fs.Fs, repo)
			assert.NoError(err)
			assert.Equal(result, tt.result)
		})
	}
}
