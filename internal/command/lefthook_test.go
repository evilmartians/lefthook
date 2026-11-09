package command

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestReloadConfigDoesNotDuplicateJobs(t *testing.T) {
	for _, tt := range []struct {
		name, primary, override string
		withLocal               bool
	}{
		{name: "local fallback", primary: "lefthook-local.yml"},
		{name: "local override", primary: ".lefthook-local.yml", override: ".lefthook-local.yml"},
		{name: "main config", primary: "lefthook.yml"},
		{name: "main and local configs", primary: "lefthook.yml", withLocal: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, err := filepath.Abs("src")
			require.NoError(t, err)
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			require.NoError(t, fs.MkdirAll(root, 0o755))
			require.NoError(t, fs.WriteFile(filepath.Join(root, tt.primary), []byte(`
extends: [base.yml]
remotes:
  - git_url: https://example.com/hooks
    ref: missing-ref
pre-commit:
  jobs:
    - run: echo primary
`), 0o644))
			require.NoError(t, fs.WriteFile(filepath.Join(root, "base.yml"), []byte(`
pre-commit:
  jobs:
    - run: echo base
`), 0o644))
			if tt.withLocal {
				require.NoError(t, fs.WriteFile(filepath.Join(root, "lefthook-local.yml"), []byte(`
pre-commit:
  jobs:
    - run: echo local
`), 0o644))
			}
			t.Setenv("LEFTHOOK_CONFIG", tt.override)
			repo := gittest.NewRepositoryBuilder().Fs(fs).Root(root).Build()
			l := &Lefthook{fs: fs, repo: repo, logger: loggertest.New()}
			cfg, err := l.LoadConfig()
			require.NoError(t, err)
			require.Len(t, cfg.Remotes, 1)
			want := []string{"echo primary", "echo base"}
			if tt.withLocal {
				want = append(want, "echo local")
			}
			assert.Len(t, cfg.Hooks["pre-commit"].Jobs, len(want))

			// A failed sync can select an already fetched fallback ref.
			cfg.Remotes[0].Ref = "fallback-ref"
			remotePath := repo.RemoteFolder(cfg.Remotes[0].GitURL, cfg.Remotes[0].Ref)
			require.NoError(t, fs.MkdirAll(remotePath, 0o755))
			require.NoError(t, fs.WriteFile(filepath.Join(remotePath, "lefthook.yml"), []byte(`
pre-commit:
  jobs:
    - run: echo remote
`), 0o644))

			cfg, err = l.reloadConfig(cfg)
			require.NoError(t, err)
			commands := make([]string, 0, len(cfg.Hooks["pre-commit"].Jobs))
			for _, job := range cfg.Hooks["pre-commit"].Jobs {
				commands = append(commands, job.Run)
			}
			assert.ElementsMatch(t, append(want, "echo remote"), commands)
			assert.Equal(t, "fallback-ref", cfg.Remotes[0].Ref)
			checksum, err := cfg.Md5()
			require.NoError(t, err)
			reloaded, err := l.reloadConfig(cfg)
			require.NoError(t, err)
			reloadedChecksum, err := reloaded.Md5()
			require.NoError(t, err)
			assert.Equal(t, checksum, reloadedChecksum)
		})
	}
}
