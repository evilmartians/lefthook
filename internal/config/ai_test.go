package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestAIConfig(t *testing.T) {
	root, err := filepath.Abs("")
	assert.NoError(t, err)

	for name, tt := range map[string]struct {
		files  map[string]string
		result *Config
	}{
		"with claude hooks": {
			files: map[string]string{
				"lefthook.yml": `
ai:
  claude:
    Stop: validate
    PreToolUse: security-check

validate:
  jobs:
    - run: go test ./...
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				AI: &AI{
					Claude: map[string]string{
						"Stop":       "validate",
						"PreToolUse": "security-check",
					},
				},
				Hooks: map[string]*Hook{
					"validate": {
						Name: "validate",
						Jobs: []*Job{
							{Run: "go test ./..."},
						},
					},
				},
			},
		},
		"with codex hooks": {
			files: map[string]string{
				"lefthook.yml": `
ai:
  codex:
    Stop: validate

validate:
  jobs:
    - run: go test ./...
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				AI: &AI{
					Codex: map[string]string{
						"Stop": "validate",
					},
				},
				Hooks: map[string]*Hook{
					"validate": {
						Name: "validate",
						Jobs: []*Job{
							{Run: "go test ./..."},
						},
					},
				},
			},
		},
		"with both claude and codex hooks": {
			files: map[string]string{
				"lefthook.yml": `
ai:
  claude:
    Stop: validate
  codex:
    PreToolUse: security-check

validate:
  jobs:
    - run: go test ./...

security-check:
  jobs:
    - run: ./scripts/security.sh
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				AI: &AI{
					Claude: map[string]string{
						"Stop": "validate",
					},
					Codex: map[string]string{
						"PreToolUse": "security-check",
					},
				},
				Hooks: map[string]*Hook{
					"validate": {
						Name: "validate",
						Jobs: []*Job{
							{Run: "go test ./..."},
						},
					},
					"security-check": {
						Name: "security-check",
						Jobs: []*Job{
							{Run: "./scripts/security.sh"},
						},
					},
				},
			},
		},
		"without ai section": {
			files: map[string]string{
				"lefthook.yml": `
pre-commit:
  jobs:
    - run: go vet ./...
`,
			},
			result: &Config{
				SourceDir:      DefaultSourceDir,
				SourceDirLocal: DefaultSourceDirLocal,
				Hooks: map[string]*Hook{
					"pre-commit": {
						Name: "pre-commit",
						Jobs: []*Job{
							{Run: "go vet ./..."},
						},
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("TestAIConfig/%s", name), func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
			loader := NewLoader(repo, loggertest.New())

			for path, content := range tt.files {
				fullPath := filepath.Join(root, path)
				assert.NoError(fs.WriteFile(fullPath, []byte(content), 0o644))
				defer func() { _ = fs.Remove(fullPath) }()
			}

			result, err := loader.Load()
			assert.NoError(err)
			assert.Equal(tt.result, result)
		})
	}
}
