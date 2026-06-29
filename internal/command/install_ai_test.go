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

	claudePath := func() string {
		return filepath.Join(root, claudeSettingsDir, claudeSettingsFile)
	}

	codexPath := func() string {
		return filepath.Join(root, codexHooksDir, codexHooksFile)
	}

	cursorPath := func() string {
		return filepath.Join(root, cursorHooksDir, cursorHooksFile)
	}

	for n, tt := range map[string]struct {
		ai            *config.AI
		existingFiles map[string]string
		wantClaude    map[string]any
		wantCodex     map[string]any
		wantCursor    map[string]any
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
		"cursor only - creates hooks.json": {
			ai: &config.AI{
				Cursor: map[string]string{
					"stop": "validate",
				},
			},
			wantCursor: map[string]any{
				"version": float64(cursorHooksVersion),
				"hooks": map[string]any{
					"stop": []any{
						map[string]any{
							"command": "lefthook run validate",
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
		"cursor merges with existing settings preserving non-lefthook entries": {
			ai: &config.AI{
				Cursor: map[string]string{
					"stop": "validate",
				},
			},
			existingFiles: map[string]string{
				cursorPath(): `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "./custom.sh" }
    ]
  }
}`,
			},
			wantCursor: map[string]any{
				"version": float64(cursorHooksVersion),
				"hooks": map[string]any{
					"stop": []any{
						map[string]any{"command": "./custom.sh"},
						map[string]any{"command": "lefthook run validate"},
					},
				},
			},
		},
		"cursor replaces stale lefthook entries on re-install": {
			ai: &config.AI{
				Cursor: map[string]string{
					"stop": "validate",
				},
			},
			existingFiles: map[string]string{
				cursorPath(): `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "lefthook run old-hook" }
    ]
  }
}`,
			},
			wantCursor: map[string]any{
				"version": float64(cursorHooksVersion),
				"hooks": map[string]any{
					"stop": []any{
						map[string]any{"command": "lefthook run validate"},
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

			err := l.installAIHooks(tt.ai, &config.Config{Lefthook: "lefthook"})
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

			if tt.wantCursor != nil {
				data, readErr := afero.ReadFile(fs, cursorPath())
				assert.NoError(readErr)

				var got map[string]any
				assert.NoError(json.Unmarshal(data, &got))
				assert.Equal(tt.wantCursor, got)
			} else {
				exists, _ := afero.Exists(fs, cursorPath())
				assert.False(exists, "cursor hooks file should not exist")
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

func TestIsLefthookRunCommand(t *testing.T) {
	for n, tt := range map[string]struct {
		cmd  string
		want bool
	}{
		"legacy name":            {cmd: "lefthook run lint", want: true},
		"absolute path":          {cmd: "/home/user/repo/lefthook run lint", want: true},
		"quoted path with space": {cmd: "'/home/user/my repo/lefthook' run lint", want: true},
		"windows path":           {cmd: `C:\tools\lefthook.exe run lint`, want: true},
		"user command":           {cmd: "./custom.sh lint", want: false},
		"other binary":           {cmd: "/usr/bin/other run lint", want: false},
	} {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, tt.want, isLefthookRunCommand(tt.cmd))
		})
	}
}

func TestLefthookRunCommand(t *testing.T) {
	assert.Equal(t, "lefthook run lint", lefthookRunCommand("lefthook", "lint", false))
	assert.Equal(t, "'/my path/lefthook' run lint", lefthookRunCommand("/my path/lefthook", "lint", true))
}

func TestValidateAIHooks(t *testing.T) {
	hooks := map[string]*config.Hook{
		"validate":       {Name: "validate"},
		"security-check": {Name: "security-check"},
	}

	for n, tt := range map[string]struct {
		ai      *config.AI
		wantErr error
	}{
		"all references resolve": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "validate"},
				Codex:  map[string]string{"PreToolUse": "security-check"},
				Cursor: map[string]string{"stop": "validate"},
			},
		},
		"missing claude reference": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "no-such-hook"},
			},
			wantErr: errAIHooksMisconfigured,
		},
		"missing codex reference": {
			ai: &config.AI{
				Codex: map[string]string{"Stop": "missing"},
			},
			wantErr: errAIHooksMisconfigured,
		},
		"missing cursor reference": {
			ai: &config.AI{
				Cursor: map[string]string{"stop": "missing"},
			},
			wantErr: errAIHooksMisconfigured,
		},
		"multiple missing references are sorted": {
			ai: &config.AI{
				Claude: map[string]string{"Stop": "zzz"},
				Codex:  map[string]string{"Stop": "aaa"},
			},
			wantErr: errAIHooksMisconfigured,
		},
	} {
		t.Run(n, func(t *testing.T) {
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
	codexPath := filepath.Join(root, codexHooksDir, codexHooksFile)
	cursorPath := filepath.Join(root, cursorHooksDir, cursorHooksFile)

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

	const cursorLefthookOnly = `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "lefthook run validate" }
    ]
  }
}`

	const cursorMixed = `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "lefthook run validate" },
      { "command": "./custom.sh" }
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
			wantRemoved:   []string{claudePath, codexPath, cursorPath},
		},
		"removes cursor file with only lefthook entries": {
			existingFiles: map[string]string{cursorPath: cursorLefthookOnly},
			wantRemoved:   []string{cursorPath},
		},
		"preserves cursor user entries": {
			existingFiles: map[string]string{cursorPath: cursorMixed},
			wantContent: map[string]map[string]any{
				cursorPath: {
					"version": float64(1),
					"hooks": map[string]any{
						"stop": []any{
							map[string]any{"command": "./custom.sh"},
						},
					},
				},
			},
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
