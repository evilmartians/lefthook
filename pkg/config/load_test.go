package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
)

type testcase struct {
	global []byte
	local  []byte
	result *Config
}

func TestLoad(t *testing.T) {
	testCases := [...]testcase{
		testcase{
			global: []byte(`
min_version: 0.6.0
colors: false
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
colors: true

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
				Colors:         true,
				SourceDir:      "$HOME/sources",
				SourceDirLocal: "$HOME/sources_local",

				Hooks: map[string]*Hook{
					"pre-commit": &Hook{
						Glob:     "*.rb",
						Parallel: true,
						Commands: map[string]*Command{
							"tests": &Command{
								Skip: true,
								Run:  "bundle exec rspec",
								Tags: []string{"backend", "test"},
							},
							"lint": &Command{
								Skip: false,
								Run:  "docker exec -it ruby:2.7 bundle exec rubocop",
								Tags: []string{"backend", "linter"},
							},
						},
						Scripts: map[string]*Script{
							"format.sh": &Script{
								Skip:   true,
								Runner: "bash",
							},
						},
					},
					"pre-push": &Hook{
						Commands: map[string]*Command{
							"rubocop": &Command{
								Run:  "bundle exec rubocop",
								Tags: []string{"backend", "linter"},
							},
						},
					},
				},
			},
		},
	}

	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	for _, tc := range testCases {
		afs.WriteFile("/lefthook.yml", tc.global, 0644)
		afs.WriteFile("/lefthook-local.yml", tc.local, 0644)

		checkConfig, err := Load(fs, "/")

		if err != nil {
			t.Errorf("should parse configs without errors: %s", err)
		} else {
			if !cmp.Equal(checkConfig, tc.result, cmpopts.IgnoreUnexported(Hook{})) {
				t.Errorf("configs should be equal")
				t.Errorf("(-want +got):\n%s", cmp.Diff(tc.result, checkConfig))
			}
		}
	}
}
