package command

import (
	"encoding/json"
	"fmt"
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

	claudePath := func() string {
		return filepath.Join(root, claudeSettingsDir, claudeSettingsFile)
	}

	codexPath := func() string {
		return filepath.Join(root, codexHooksDir, codexHooksFile)
	}

	for n, tt := range map[string]struct {
		ai            *config.AI
		existingFiles map[string]string
		wantClaude    map[string]any
		wantCodex     map[string]any
	}{
		"claude only - creates settings.json": {
			ai: &config.AI{
				Claude: map[string]string{
					"Stop": "validate",
				},
			},
			wantClaude: map[string]any{
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run validate",
								},
							},
						},
					},
				},
			},
		},
		"codex only - creates hooks.json": {
			ai: &config.AI{
				Codex: map[string]string{
					"PreToolUse": "security-check",
				},
			},
			wantCodex: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run security-check",
								},
							},
						},
					},
				},
			},
		},
		"both providers": {
			ai: &config.AI{
				Claude: map[string]string{
					"Stop": "validate",
				},
				Codex: map[string]string{
					"Stop": "validate",
				},
			},
			wantClaude: map[string]any{
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run validate",
								},
							},
						},
					},
				},
			},
			wantCodex: map[string]any{
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run validate",
								},
							},
						},
					},
				},
			},
		},
		"merges with existing settings preserving non-lefthook entries": {
			ai: &config.AI{
				Claude: map[string]string{
					"Stop": "validate",
				},
			},
			existingFiles: map[string]string{
				claudePath(): `{
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
			},
			wantClaude: map[string]any{
				"permissions": map[string]any{"allow": []any{"Bash"}},
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo 'user hook'",
								},
							},
						},
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run validate",
								},
							},
						},
					},
				},
			},
		},
		"replaces stale lefthook entries on re-install": {
			ai: &config.AI{
				Claude: map[string]string{
					"Stop": "validate",
				},
			},
			existingFiles: map[string]string{
				claudePath(): `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "lefthook run old-hook" }
        ]
      }
    ]
  }
}`,
			},
			wantClaude: map[string]any{
				"hooks": map[string]any{
					"Stop": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "lefthook run validate",
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("TestInstallAIHooks/%s", n), func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.NewMemMapFs()
			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
			l := &Lefthook{
				logger: loggertest.New(),
				fs:     fs,
				repo:   repo,
			}

			for path, content := range tt.existingFiles {
				dir := filepath.Dir(path)
				assert.NoError(fs.MkdirAll(dir, hooksDirMode))
				assert.NoError(afero.WriteFile(fs, path, []byte(content), checksumFileMode))
			}

			err := l.installAIHooks(tt.ai)
			assert.NoError(err)

			if tt.wantClaude != nil {
				data, readErr := afero.ReadFile(fs, claudePath())
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(tt.wantClaude, got)
			} else {
				exists, _ := afero.Exists(fs, claudePath())
				assert.False(exists, "claude settings file should not exist")
			}

			if tt.wantCodex != nil {
				data, readErr := afero.ReadFile(fs, codexPath())
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(tt.wantCodex, got)
			} else {
				exists, _ := afero.Exists(fs, codexPath())
				assert.False(exists, "codex hooks file should not exist")
			}
		})
	}
}

func TestValidateAIHooks(t *testing.T) {
	hooks := map[string]*config.Hook{
		"validate":       {Name: "validate"},
		"security-check": {Name: "security-check"},
	}

	for n, tt := range map[string]struct {
		ai      *config.AI
		wantErr string
	}{
		"all references resolve": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "validate"},
				Codex:  map[string]string{"PreToolUse": "security-check"},
			},
		},
		"missing claude reference": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "no-such-hook"},
			},
			wantErr: `ai config references undefined hooks: ai.claude.Stop -> "no-such-hook"`,
		},
		"missing codex reference": {
			ai: &config.AI{
				Codex: map[string]string{"Stop": "missing"},
			},
			wantErr: `ai config references undefined hooks: ai.codex.Stop -> "missing"`,
		},
		"multiple missing references are sorted": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "zzz"},
				Codex:  map[string]string{"Stop": "aaa"},
			},
			wantErr: `ai config references undefined hooks: ai.claude.Stop -> "zzz", ai.codex.Stop -> "aaa"`,
		},
	} {
		t.Run(n, func(t *testing.T) {
			assert := assert.New(t)

			err := validateAIHooks(tt.ai, hooks)
			if tt.wantErr == "" {
				assert.NoError(err)
			} else {
				assert.EqualError(err, tt.wantErr)
			}
		})
	}
}

func TestUninstallAIHooks(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	claudePath := filepath.Join(root, claudeSettingsDir, claudeSettingsFile)
	codexPath := filepath.Join(root, codexHooksDir, codexHooksFile)

	const lefthookOnly = `{
  "hooks": {
    "Stop": [
      { "hooks": [{ "type": "command", "command": "lefthook run validate" }] }
    ]
  }
}`

	const mixed = `{
  "model": "sonnet",
  "hooks": {
    "Stop": [
      { "hooks": [{ "type": "command", "command": "lefthook run validate" }] },
      { "hooks": [{ "type": "command", "command": "./custom.sh" }] }
    ]
  }
}`

	for n, tt := range map[string]struct {
		existingFiles map[string]string
		wantRemoved   []string
		wantContent   map[string]map[string]any
	}{
		"removes file with only lefthook entries": {
			existingFiles: map[string]string{claudePath: lefthookOnly},
			wantRemoved:   []string{claudePath},
		},
		"preserves user entries and top-level keys": {
			existingFiles: map[string]string{claudePath: mixed},
			wantContent: map[string]map[string]any{
				claudePath: {
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
				},
			},
		},
		"handles missing files": {
			existingFiles: map[string]string{},
			wantRemoved:   []string{claudePath, codexPath},
		},
		"handles empty files": {
			existingFiles: map[string]string{
				claudePath: "",
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
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

			assert.NoError(l.uninstallAIHooks())

			for _, path := range tt.wantRemoved {
				exists, _ := afero.Exists(fs, path)
				assert.False(exists, "%s should be removed", path)
			}

			for path, want := range tt.wantContent {
				data, readErr := afero.ReadFile(fs, path)
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(want, got)
			}
		})
	}
}
