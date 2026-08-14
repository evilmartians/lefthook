package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestInstallAIHooks(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	paths := map[string]string{
		"claude":  filepath.Join(root, claudeSettingsDir, claudeSettingsFile),
		"codex":   filepath.Join(root, codexHooksDir, codexHooksFile),
		"cursor":  filepath.Join(root, cursorHooksDir, cursorHooksFile),
		"copilot": filepath.Join(root, copilotHooksDir, copilotHooksFile),
	}

	for name, tt := range map[string]struct {
		ai            *config.AI
		existingFiles map[string]string
		wantFiles     map[string]map[string]any
		wantMissing   []string
	}{
		"writes configured providers": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "validate"},
				Codex:  map[string]string{"PreToolUse": "security-check"},
			},
			wantFiles: map[string]map[string]any{
				paths["claude"]: {
					"hooks": map[string]any{
						"Stop": []any{
							map[string]any{
								"hooks": []any{
									map[string]any{"type": "command", "command": "lefthook run validate"},
								},
							},
						},
					},
				},
				paths["codex"]: {
					"hooks": map[string]any{
						"PreToolUse": []any{
							map[string]any{
								"hooks": []any{
									map[string]any{"type": "command", "command": "lefthook run security-check"},
								},
							},
						},
					},
				},
			},
			wantMissing: []string{paths["cursor"], paths["copilot"]},
		},
		"merges existing claude and cursor hooks": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "validate"},
				Cursor: map[string]string{"stop": "validate"},
			},
			existingFiles: map[string]string{
				paths["claude"]: `{
  "permissions": { "allow": ["Bash"] },
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "echo 'user hook'" }
        ]
      }
    ]
  }
}`,
				paths["cursor"]: `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "./custom.sh" }
    ]
  }
}`,
			},
			wantFiles: map[string]map[string]any{
				paths["claude"]: {
					"permissions": map[string]any{"allow": []any{"Bash"}},
					"hooks": map[string]any{
						"Stop": []any{
							map[string]any{
								"hooks": []any{
									map[string]any{"type": "command", "command": "echo 'user hook'"},
								},
							},
							map[string]any{
								"hooks": []any{
									map[string]any{"type": "command", "command": "lefthook run validate"},
								},
							},
						},
					},
				},
				paths["cursor"]: {
					"version": float64(cursorHooksVersion),
					"hooks": map[string]any{
						"stop": []any{
							map[string]any{"command": "./custom.sh"},
							map[string]any{"command": "lefthook run validate"},
						},
					},
				},
			},
			wantMissing: []string{paths["codex"], paths["copilot"]},
		},
		"overwrites copilot file completely": {
			ai: &config.AI{
				Copilot: map[string]string{"postToolUse": "validate"},
			},
			existingFiles: map[string]string{
				paths["copilot"]: `{"version":1,"hooks":{"postToolUse":[{"command":"./custom.sh"}]}}`,
			},
			wantFiles: map[string]map[string]any{
				paths["copilot"]: {
					"version": float64(copilotHooksVersion),
					"hooks": map[string]any{
						"postToolUse": []any{
							map[string]any{"command": "lefthook run validate"},
						},
					},
				},
			},
			wantMissing: []string{paths["claude"], paths["codex"], paths["cursor"]},
		},
	} {
		t.Run(fmt.Sprintf("TestInstallAIHooks/%s", name), func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.NewMemMapFs()
			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
			l := &Lefthook{
				logger: loggertest.New(),
				fs:     fs,
				repo:   repo,
			}

			for path, content := range tt.existingFiles {
				assert.NoError(fs.MkdirAll(filepath.Dir(path), hooksDirMode))
				assert.NoError(afero.WriteFile(fs, path, []byte(content), checksumFileMode))
			}

			err := l.installAIHooks(tt.ai, &config.Config{Lefthook: "lefthook"})
			assert.NoError(err)

			for path, want := range tt.wantFiles {
				data, readErr := afero.ReadFile(fs, path)
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(want, got)
			}

			for _, path := range tt.wantMissing {
				exists, existsErr := afero.Exists(fs, path)
				assert.NoError(existsErr)
				assert.False(exists, "%s should not exist", path)
			}
		})
	}
}

func TestResolveLefthookBin(t *testing.T) {
	t.Run("uses config lefthook setting", func(t *testing.T) {
		bin, quote := resolveLefthookBin(&config.Config{Lefthook: "bundle exec lefthook"})
		assert.Equal(t, "bundle exec lefthook", bin)
		assert.False(t, quote)
	})

	t.Run("falls back to lefthook name", func(t *testing.T) {
		bin, quote := resolveLefthookBin(nil)
		if _, err := os.Executable(); err != nil {
			assert.Equal(t, "lefthook", bin)
			assert.False(t, quote)
			return
		}

		assert.NotEmpty(t, bin)
	})
}

func TestLefthookRunCommand(t *testing.T) {
	assert.Equal(t, "lefthook run lint", lefthookRunCommand("lefthook", "lint", false))
	assert.Equal(t, "'/my path/lefthook' run lint", lefthookRunCommand("/my path/lefthook", "lint", true))
}

func TestLefthookDetection(t *testing.T) {
	t.Run("flat hook detects lefthook only", func(t *testing.T) {
		assert.True(t, isFlatLefthookEntry(map[string]any{"command": "lefthook run lint"}))
		assert.True(t, isFlatLefthookEntry(map[string]any{"command": "/tmp/bin/lefthook run lint"}))
		assert.False(t, isFlatLefthookEntry(map[string]any{"command": "npm run lint"}))
	})

	t.Run("matcher detects lefthook only", func(t *testing.T) {
		assert.True(t, isLefthookMatcher(map[string]any{
			"hooks": []any{
				map[string]any{"command": "lefthook run lint"},
			},
		}))
		assert.False(t, isLefthookMatcher(map[string]any{
			"hooks": []any{
				map[string]any{"command": "npm run lint"},
			},
		}))
	})
}

func TestValidateAIHooks(t *testing.T) {
	hooks := map[string]*config.Hook{
		"validate":       {Name: "validate"},
		"security-check": {Name: "security-check"},
	}

	for name, tt := range map[string]struct {
		ai      *config.AI
		wantErr error
	}{
		"all references resolve": {
			ai: &config.AI{
				Claude:  map[string]string{"Stop": "validate"},
				Codex:   map[string]string{"PreToolUse": "security-check"},
				Cursor:  map[string]string{"stop": "validate"},
				Copilot: map[string]string{"postToolUse": "validate"},
			},
		},
		"missing claude reference": {
			ai:      &config.AI{Claude: map[string]string{"Stop": "missing"}},
			wantErr: errAIHooksMisconfigured,
		},
		"missing codex reference": {
			ai:      &config.AI{Codex: map[string]string{"Stop": "missing"}},
			wantErr: errAIHooksMisconfigured,
		},
		"missing cursor reference": {
			ai:      &config.AI{Cursor: map[string]string{"stop": "missing"}},
			wantErr: errAIHooksMisconfigured,
		},
		"missing copilot reference": {
			ai:      &config.AI{Copilot: map[string]string{"postToolUse": "missing"}},
			wantErr: errAIHooksMisconfigured,
		},
		"multiple missing references are sorted": {
			ai: &config.AI{
				Claude:  map[string]string{"Stop": "zzz"},
				Codex:   map[string]string{"Stop": "aaa"},
				Copilot: map[string]string{"postToolUse": "mmm"},
			},
			wantErr: errAIHooksMisconfigured,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			l := &Lefthook{logger: loggertest.New()}
			err := l.validateAIHooks(tt.ai, hooks)
			if tt.wantErr == nil {
				assert.NoError(err)
			} else {
				assert.ErrorIs(err, tt.wantErr)
			}
		})
	}
}

func TestUninstallAIHooks(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	claudePath := filepath.Join(root, claudeSettingsDir, claudeSettingsFile)
	cursorPath := filepath.Join(root, cursorHooksDir, cursorHooksFile)
	copilotPath := filepath.Join(root, copilotHooksDir, copilotHooksFile)

	t.Run("strips managed claude and cursor hooks but deletes copilot file", func(t *testing.T) {
		assert := assert.New(t)

		fs := afero.NewMemMapFs()
		repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
		l := &Lefthook{
			logger: loggertest.New(),
			fs:     fs,
			repo:   repo,
		}

		assert.NoError(fs.MkdirAll(filepath.Dir(claudePath), hooksDirMode))
		assert.NoError(afero.WriteFile(fs, claudePath, []byte(`{
  "model": "sonnet",
  "hooks": {
    "Stop": [
      { "hooks": [{ "type": "command", "command": "lefthook run validate" }] },
      { "hooks": [{ "type": "command", "command": "./custom.sh" }] }
    ]
  }
}`), checksumFileMode))
		assert.NoError(fs.MkdirAll(filepath.Dir(cursorPath), hooksDirMode))
		assert.NoError(afero.WriteFile(fs, cursorPath, []byte(`{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "lefthook run validate" },
      { "command": "./custom.sh" }
    ]
  }
}`), checksumFileMode))
		assert.NoError(fs.MkdirAll(filepath.Dir(copilotPath), hooksDirMode))
		assert.NoError(afero.WriteFile(fs, copilotPath, []byte(`{invalid json`), checksumFileMode))

		assert.NoError(l.uninstallAIHooks())

		claudeData, readErr := afero.ReadFile(fs, claudePath)
		assert.NoError(readErr)
		var claude map[string]any
		assert.NoError(json.Unmarshal(claudeData, &claude))
		assert.Equal(map[string]any{
			"model": "sonnet",
			"hooks": map[string]any{
				"Stop": []any{
					map[string]any{
						"hooks": []any{
							map[string]any{"type": "command", "command": "./custom.sh"},
						},
					},
				},
			},
		}, claude)

		cursorData, readErr := afero.ReadFile(fs, cursorPath)
		assert.NoError(readErr)
		var cursor map[string]any
		assert.NoError(json.Unmarshal(cursorData, &cursor))
		assert.Equal(map[string]any{
			"version": float64(1),
			"hooks": map[string]any{
				"stop": []any{
					map[string]any{"command": "./custom.sh"},
				},
			},
		}, cursor)

		exists, existsErr := afero.Exists(fs, copilotPath)
		assert.NoError(existsErr)
		assert.False(exists, "%s should be removed", copilotPath)
	})

	t.Run("ignores missing files", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
		l := &Lefthook{
			logger: loggertest.New(),
			fs:     fs,
			repo:   repo,
		}

		assert.NoError(t, l.uninstallAIHooks())
	})
}
