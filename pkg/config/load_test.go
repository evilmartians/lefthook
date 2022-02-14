package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
)

func TestLoad(t *testing.T) {
	for i, tt := range [...]struct {
		name   string
		global []byte
		local  []byte
		result *Config
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
						Glob:     "",
						Parallel: false,
						Commands: map[string]*Command{
							"tests": {
								Run:    "yarn test", // copies Runner to Run
								Runner: "yarn test",
							},
						},
					},
					"post-commit": {
						Glob:     "",
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
  glob: "*.rb"
  parallel: true
  commands:
    tests:
      run: bundle exec rspec
      tags: [backend, test]
    lint:
      run: bundle exec rubocop
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
						Glob:     "*.rb",
						Parallel: true,
						Commands: map[string]*Command{
							"tests": {
								Skip: true,
								Run:  "bundle exec rspec",
								Tags: []string{"backend", "test"},
							},
							"lint": {
								Skip: false,
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
	} {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			if err := fs.WriteFile("/lefthook.yml", tt.global, 0644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if err := fs.WriteFile("/lefthook-local.yml", tt.local, 0644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			checkConfig, err := Load(fs.Fs, "/")

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
