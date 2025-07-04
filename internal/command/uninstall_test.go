package command

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
)

func TestLefthookUninstall(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	configPath := filepath.Join(root, "lefthook.yml")
	gitPath := filepath.Join(root, ".git")
	hooksPath := filepath.Join(gitPath, "hooks")
	infoPath := filepath.Join(gitPath, "info")
	checksumPath := filepath.Join(infoPath, config.ChecksumFileName)

	hookPath := func(hook string) string {
		return filepath.Join(root, ".git", "hooks", hook)
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
				configPath,
				hookPath("pre-commit"),
			},
			wantNotExist: []string{
				checksumPath,
				hookPath("post-commit"),
			},
		},
		{
			name: "with force",
			args: UninstallArgs{Force: true},
			existingHooks: map[string]string{
				"pre-commit":  "not a lefthook hook",
				"post-commit": "\n# LEFTHOOK file\n",
			},
			config:    "# empty",
			wantExist: []string{configPath},
			wantNotExist: []string{
				checksumPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
			},
		},
		{
			name: "with --remove-config option",
			args: UninstallArgs{RemoveConfig: true},
			existingHooks: map[string]string{
				"pre-commit":  "not a lefthook hook",
				"post-commit": "# LEFTHOOK",
			},
			config: "# empty",
			wantExist: []string{
				hookPath("pre-commit"),
			},
			wantNotExist: []string{
				checksumPath,
				configPath,
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
				configPath,
				hookPath("pre-commit"),
				hookPath("post-commit"),
			},
			wantNotExist: []string{
				checksumPath,
				hookPath("post-commit.old"),
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{
				fs: fs,
				repo: &git.Repository{
					Fs:        fs,
					HooksPath: hooksPath,
					RootPath:  root,
					GitPath:   gitPath,
					InfoPath:  infoPath,
				},
			}

			// Create config and checksum file
			err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			err = afero.WriteFile(fs, checksumPath, []byte("CHECKSUM"), 0o644)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// Prepare files that should exist
			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err = fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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
