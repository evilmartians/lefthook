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

	paths := map[string]string{
		"claude":      filepath.Join(root, claudeSettingsDir, claudeSettingsFile),
		"codex":       filepath.Join(root, codexHooksDir, codexHooksFile),
		"cursor":      filepath.Join(root, cursorHooksDir, cursorHooksFile),
		"copilot":     filepath.Join(root, copilotHooksDir, copilotHooksFile),
		"antigravity": filepath.Join(root, antigravityHooksDir, antigravityHooksFile),
	}

	for name, tt := range map[string]struct {
		ai              *config.AI
		cfg             *config.Config
		existingFiles   map[string]string
		wantFiles       map[string]map[string]any
		wantRawFiles    map[string]string
		wantMissing     []string
		wantErr         error
		wantErrContains string
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
			wantMissing: []string{paths["cursor"], paths["copilot"], paths["antigravity"]},
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
			wantMissing: []string{paths["codex"], paths["copilot"], paths["antigravity"]},
		},
		"replaces entries with an install-time absolute path": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "validate"},
			},
			existingFiles: map[string]string{
				paths["claude"]: `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "/Users/me/project/node_modules/lefthook-darwin-arm64/bin/lefthook run validate" }
        ]
      }
    ]
  }
}`,
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
			},
			wantMissing: []string{paths["codex"], paths["cursor"], paths["copilot"], paths["antigravity"]},
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
			wantMissing: []string{paths["claude"], paths["codex"], paths["cursor"], paths["antigravity"]},
		},
		"writes antigravity tool and lifecycle events": {
			ai: &config.AI{
				Antigravity: map[string]string{
					"SessionStart":   "session-init",
					"PreToolUse":     "security-check",
					"PostToolUse":    "lint",
					"PreInvocation":  "pre-invoke",
					"PostInvocation": "post-invoke",
					"Stop":           "validate",
				},
			},
			wantRawFiles: map[string]string{
				paths["antigravity"]: `{
  "lefthook": {
    "PostInvocation": [
      {
        "command": "lefthook run post-invoke",
        "type": "command"
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "command": "lefthook run lint",
            "type": "command"
          }
        ],
        "matcher": "*"
      }
    ],
    "PreInvocation": [
      {
        "command": "lefthook run pre-invoke",
        "type": "command"
      }
    ],
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "lefthook run security-check",
            "type": "command"
          }
        ],
        "matcher": "*"
      }
    ],
    "SessionStart": [
      {
        "command": "lefthook run session-init",
        "type": "command"
      }
    ],
    "Stop": [
      {
        "command": "lefthook run validate",
        "type": "command"
      }
    ]
  }
}
`,
			},
			wantMissing: []string{paths["claude"], paths["codex"], paths["cursor"], paths["copilot"]},
		},
		"uses custom lefthook binary for antigravity": {
			ai: &config.AI{
				Antigravity: map[string]string{
					"PreToolUse": "security-check",
					"Stop":       "validate",
				},
			},
			cfg: &config.Config{Lefthook: "npx lefthook"},
			wantFiles: map[string]map[string]any{
				paths["antigravity"]: {
					"lefthook": map[string]any{
						"PreToolUse": []any{
							map[string]any{
								"matcher": "*",
								"hooks": []any{
									map[string]any{"type": "command", "command": "npx lefthook run security-check"},
								},
							},
						},
						"Stop": []any{
							map[string]any{"type": "command", "command": "npx lefthook run validate"},
						},
					},
				},
			},
			wantMissing: []string{paths["claude"], paths["codex"], paths["cursor"], paths["copilot"]},
		},
		"merges and replaces antigravity lefthook key while preserving user hooks": {
			ai: &config.AI{
				Antigravity: map[string]string{
					"PreToolUse": "security-check",
					"Stop":       "validate",
				},
			},
			existingFiles: map[string]string{
				paths["antigravity"]: `{
  "lint-checker": {
    "PostToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          { "type": "command", "command": "./scripts/lint.sh", "timeout": 10 }
        ]
      }
    ]
  },
  "lefthook": {
    "SessionStart": [
      { "type": "command", "command": "lefthook run stale-hook" }
    ],
    "Stop": [
      { "type": "command", "command": "lefthook run old-validate" }
    ]
  }
}`,
			},
			wantFiles: map[string]map[string]any{
				paths["antigravity"]: {
					"lint-checker": map[string]any{
						"PostToolUse": []any{
							map[string]any{
								"matcher": "run_command",
								"hooks": []any{
									map[string]any{
										"type":    "command",
										"command": "./scripts/lint.sh",
										"timeout": float64(10),
									},
								},
							},
						},
					},
					"lefthook": map[string]any{
						"PreToolUse": []any{
							map[string]any{
								"matcher": "*",
								"hooks": []any{
									map[string]any{"type": "command", "command": "lefthook run security-check"},
								},
							},
						},
						"Stop": []any{
							map[string]any{"type": "command", "command": "lefthook run validate"},
						},
					},
				},
			},
		},
		"empty antigravity config is a no-op": {
			ai: &config.AI{
				Antigravity: map[string]string{},
			},
			wantMissing: []string{paths["claude"], paths["codex"], paths["cursor"], paths["copilot"], paths["antigravity"]},
		},
		"empty antigravity config preserves existing hooks file": {
			ai: &config.AI{},
			existingFiles: map[string]string{
				paths["antigravity"]: `{"lint-checker":{"Stop":[{"type":"command","command":"./lint.sh"}]}}`,
			},
			wantRawFiles: map[string]string{
				paths["antigravity"]: `{"lint-checker":{"Stop":[{"type":"command","command":"./lint.sh"}]}}`,
			},
		},
		"rejects unsupported antigravity event": {
			ai: &config.AI{
				Antigravity: map[string]string{"UnknownEvent": "validate"},
			},
			wantErr:     errAIHooksMisconfigured,
			wantMissing: []string{paths["antigravity"]},
		},
		"fails on malformed antigravity hooks json": {
			ai: &config.AI{
				Antigravity: map[string]string{"Stop": "validate"},
			},
			existingFiles: map[string]string{
				paths["antigravity"]: `{invalid json`,
			},
			wantErrContains: "could not parse",
			wantRawFiles: map[string]string{
				paths["antigravity"]: `{invalid json`,
			},
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

			cfg := tt.cfg
			if cfg == nil {
				cfg = &config.Config{Lefthook: "lefthook"}
			}

			err := l.installAIHooks(tt.ai, cfg)
			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(err, tt.wantErr)
			case tt.wantErrContains != "":
				assert.ErrorContains(err, tt.wantErrContains)
			default:
				assert.NoError(err)
			}

			for path, want := range tt.wantFiles {
				data, readErr := afero.ReadFile(fs, path)
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(want, got)
			}

			for path, wantRaw := range tt.wantRawFiles {
				data, readErr := afero.ReadFile(fs, path)
				assert.NoError(readErr)
				assert.Equal(wantRaw, string(data))
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
		assert.Equal(t, "bundle exec lefthook", resolveLefthookBin(&config.Config{Lefthook: "bundle exec lefthook"}))
	})

	t.Run("falls back to lefthook name", func(t *testing.T) {
		assert.Equal(t, "lefthook", resolveLefthookBin(nil))
		assert.Equal(t, "lefthook", resolveLefthookBin(&config.Config{}))
	})
}

func TestLefthookRunCommand(t *testing.T) {
	assert.Equal(t, "lefthook run lint", lefthookRunCommand("lefthook", "lint"))
	assert.Equal(t, "pnpm exec lefthook run lint", lefthookRunCommand("pnpm exec lefthook", "lint"))
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
				Antigravity: map[string]string{
					"SessionStart":   "validate",
					"PreToolUse":     "security-check",
					"PostToolUse":    "validate",
					"PreInvocation":  "validate",
					"PostInvocation": "validate",
					"Stop":           "validate",
				},
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
		"missing antigravity reference": {
			ai:      &config.AI{Antigravity: map[string]string{"Stop": "missing"}},
			wantErr: errAIHooksMisconfigured,
		},
		"invalid antigravity event": {
			ai:      &config.AI{Antigravity: map[string]string{"UnknownEvent": "validate"}},
			wantErr: errAIHooksMisconfigured,
		},
		"lowercase antigravity event": {
			ai:      &config.AI{Antigravity: map[string]string{"stop": "validate"}},
			wantErr: errAIHooksMisconfigured,
		},
		"multiple missing references are sorted": {
			ai: &config.AI{
				Claude:      map[string]string{"Stop": "zzz"},
				Codex:       map[string]string{"Stop": "aaa"},
				Copilot:     map[string]string{"postToolUse": "mmm"},
				Antigravity: map[string]string{"Stop": "bbb"},
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
	antigravityDir := filepath.Join(root, antigravityHooksDir)
	antigravityPath := filepath.Join(antigravityDir, antigravityHooksFile)

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

	for name, tt := range map[string]struct {
		files       map[string]string
		wantFiles   map[string]map[string]any
		wantExist   []string
		wantMissing []string
	}{
		"antigravity preserves user hook groups": {
			files: map[string]string{
				antigravityPath: `{
  "lint-checker": {
    "PostToolUse": [
      {
        "matcher": "run_command",
        "hooks": [{ "type": "command", "command": "./scripts/lint.sh" }]
      }
    ]
  },
  "lefthook": {
    "Stop": [
      { "type": "command", "command": "lefthook run validate" }
    ]
  }
}`,
			},
			wantFiles: map[string]map[string]any{
				antigravityPath: {
					"lint-checker": map[string]any{
						"PostToolUse": []any{
							map[string]any{
								"matcher": "run_command",
								"hooks": []any{
									map[string]any{"type": "command", "command": "./scripts/lint.sh"},
								},
							},
						},
					},
				},
			},
		},
		"antigravity deletes hooks file and empty agents dir": {
			files: map[string]string{
				antigravityPath: `{
  "lefthook": {
    "Stop": [
      { "type": "command", "command": "lefthook run validate" }
    ]
  }
}`,
			},
			wantMissing: []string{antigravityPath, antigravityDir},
		},
		"antigravity deletes hooks file but keeps non-empty agents dir": {
			files: map[string]string{
				antigravityPath: `{
  "lefthook": {
    "Stop": [
      { "type": "command", "command": "lefthook run validate" }
    ]
  }
}`,
				filepath.Join(antigravityDir, "rules", "style.md"): "# Style\n",
			},
			wantExist:   []string{antigravityDir, filepath.Join(antigravityDir, "rules", "style.md")},
			wantMissing: []string{antigravityPath},
		},
		"antigravity leaves file without lefthook key intact": {
			files: map[string]string{
				antigravityPath: `{
  "lint-checker": {
    "Stop": [{ "type": "command", "command": "./lint.sh" }]
  }
}`,
			},
			wantFiles: map[string]map[string]any{
				antigravityPath: {
					"lint-checker": map[string]any{
						"Stop": []any{
							map[string]any{"type": "command", "command": "./lint.sh"},
						},
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("TestUninstallAIHooks/%s", name), func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.NewMemMapFs()
			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
			l := &Lefthook{
				logger: loggertest.New(),
				fs:     fs,
				repo:   repo,
			}

			for path, content := range tt.files {
				assert.NoError(fs.MkdirAll(filepath.Dir(path), hooksDirMode))
				assert.NoError(afero.WriteFile(fs, path, []byte(content), checksumFileMode))
			}

			assert.NoError(l.uninstallAIHooks())

			for path, want := range tt.wantFiles {
				data, readErr := afero.ReadFile(fs, path)
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(want, got)
			}

			for _, path := range tt.wantExist {
				exists, existsErr := afero.Exists(fs, path)
				assert.NoError(existsErr)
				assert.True(exists, "%s should exist", path)
			}

			for _, path := range tt.wantMissing {
				exists, existsErr := afero.Exists(fs, path)
				assert.NoError(existsErr)
				assert.False(exists, "%s should not exist", path)
			}
		})
	}

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
