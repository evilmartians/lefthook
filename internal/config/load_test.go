package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestLoad(t *testing.T) {
	root, err := filepath.Abs("")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for i, tt := range [...]struct {
		name             string
		global           []byte
		local            []byte
		remote           []byte
		remoteConfigPath string
		result           *Config
	}{
		{
			name: "simple",
			global: []byte(`
pre-commit:
  commands:
    tests:
      runner: yarn test # Using deprecated field
`),
			local: []byte(`
post-commit:
  commands:
    ping-done:
      run: curl -x POST status.com/done
`),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         true, // defaults to true
				Hooks: map[string]*Hook{
					"pre-commit": {
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run:    "yarn test", // copies Runner to Run
								Runner: "yarn test",
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
			global: []byte(`
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
`),
			local: []byte(`
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
      skip: true

pre-push:
  commands:
    rubocop:
      run: bundle exec rubocop
      tags: [backend, linter]
`),
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
								Skip:   true,
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
			name: "with extra hooks",
			global: []byte(`
tests:
  commands:
    tests:
      run: go test ./...

lints:
  scripts:
    "linter.sh":
      runner: bash
`),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         DefaultColorsEnabled,
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
			global: []byte(`
remote:
  git_url: git@github.com:evilmartians/lefthook
`),
			remote: []byte(`
pre-commit:
  commands:
    lint:
      run: yarn lint
  scripts:
    "test.sh":
      runner: bash
`),
			remoteConfigPath: filepath.Join(root, ".git", "info", "remotes", "lefthook", "lefthook.yml"),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         DefaultColorsEnabled,
				Remote: Remote{
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
			global: []byte(`
remote:
  git_url: git@github.com:evilmartians/lefthook
  ref: v1.0.0
  config: examples/custom.yml

pre-commit:
  commands:
    global:
      run: echo 'Global!'
    lint:
      run: this will be overwritten
`),
			remote: []byte(`
pre-commit:
  commands:
    lint:
      run: yarn lint
      skip: true
  scripts:
    "test.sh":
      runner: bash
`),
			remoteConfigPath: filepath.Join(root, ".git", "info", "remotes", "lefthook", "examples", "custom.yml"),
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Colors:         DefaultColorsEnabled,
				Remote: Remote{
					GitURL: "git@github.com:evilmartians/lefthook",
					Ref:    "v1.0.0",
					Config: "examples/custom.yml",
				},
				Hooks: map[string]*Hook{
					"pre-commit": {
						Commands: map[string]*Command{
							"lint": {
								Run:  "yarn lint",
								Skip: true,
							},
							"global": {
								Run: "echo 'Global!'",
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
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		repo := &git.Repository{
			Fs:       fs,
			RootPath: root,
			InfoPath: filepath.Join(root, ".git", "info"),
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			if err := fs.WriteFile(filepath.Join(root, "lefthook.yml"), tt.global, 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if err := fs.WriteFile(filepath.Join(root, "lefthook-local.yml"), tt.local, 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if len(tt.remoteConfigPath) > 0 {
				if err := fs.MkdirAll(filepath.Base(tt.remoteConfigPath), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if err := fs.WriteFile(tt.remoteConfigPath, tt.remote, 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			checkConfig, err := Load(fs.Fs, repo)

			if err != nil {
				t.Errorf("should parse configs without errors: %s", err)
			} else {
				if !cmp.Equal(checkConfig, tt.result, cmpopts.IgnoreUnexported(Hook{})) {
					t.Errorf("configs should be equal")
					t.Errorf("(-want +got):\n%s", cmp.Diff(tt.result, checkConfig))
				}
			}
		})
	}
}
