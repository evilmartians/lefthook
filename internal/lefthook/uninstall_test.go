package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestLefthookUninstall(t *testing.T) {
	repo := &git.Repository{
		HooksPath: hooksPath,
		RootPath:  root,
	}

	for n, tt := range [...]struct {
		name, config            string
		args                    UninstallArgs
		existingHooks           map[string]string
		wantExist, wantNotExist []string
	}{
		{
			name: "simple defaults",
			existingHooks: map[string]string{
				"pre-commit":  "not a lefthook hook",
				"post-commit": `"$LEFTHOOK" file`,
			},
			config: "# empty",
			wantExist: []string{
				hookPath("pre-commit"),
			},
			wantNotExist: []string{
				configPath,
				hookPath("post-commit"),
			},
		},
		{
			name: "with aggressive mode",
			args: UninstallArgs{Aggressive: true},
			existingHooks: map[string]string{
				"pre-commit":  "not a lefthook hook",
				"post-commit": "\n# LEFTHOOK file\n",
			},
			config:    "# empty",
			wantExist: []string{},
			wantNotExist: []string{
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
			},
		},
		{
			name: "with keep config arg",
			args: UninstallArgs{KeepConfiguration: true},
			existingHooks: map[string]string{
				"pre-commit":  "not a lefthook hook",
				"post-commit": "# LEFTHOOK",
			},
			config: "# empty",
			wantExist: []string{
				configPath,
				hookPath("pre-commit"),
			},
			wantNotExist: []string{
				hookPath("post-commit"),
			},
		},
		{
			name: "with .old files",
			existingHooks: map[string]string{
				"pre-commit":      "not a lefthook hook",
				"post-commit":     "LEFTHOOK file",
				"post-commit.old": "not a lefthook hook",
			},
			config: "# empty",
			wantExist: []string{
				hookPath("pre-commit"),
				hookPath("post-commit"),
			},
			wantNotExist: []string{
				configPath,
				hookPath("post-commit.old"),
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// Prepare files that should exist
			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err = fs.MkdirAll(filepath.Base(path), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err = afero.WriteFile(fs, path, []byte(content), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Do uninstall
			err = lefthook.Uninstall(&tt.args)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// Test files that should exist
			for _, file := range tt.wantExist {
				ok, err := afero.Exists(fs, file)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !ok {
					t.Errorf("expected %s to exist", file)
				}
			}

			// Test files that should not exist
			for _, file := range tt.wantNotExist {
				ok, err := afero.Exists(fs, file)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if ok {
					t.Errorf("expected %s to not exist", file)
				}
			}
		})
	}
}
