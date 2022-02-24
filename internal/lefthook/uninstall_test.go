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
		HooksPath: "/src/.git/hooks",
		RootPath:  "/src/",
	}

	for n, tt := range [...]struct {
		name                    string
		args                    UninstallArgs
		existingFiles           map[string]string
		wantExist, wantNotExist []string
	}{
		{
			name: "simple defaults",
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":  "not a lefthook hook",
				"/src/.git/hooks/post-commit": `"$LEFTHOOK" file`,
				"/src/lefthook.yml":           "# empty",
			},
			wantExist: []string{
				"/src/.git/hooks/pre-commit",
			},
			wantNotExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/post-commit",
			},
		},
		{
			name: "with aggressive mode",
			args: UninstallArgs{Aggressive: true},
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":  "not a lefthook hook",
				"/src/.git/hooks/post-commit": "\n# LEFTHOOK file\n",
				"/src/lefthook.yaml":          "# empty",
			},
			wantExist: []string{},
			wantNotExist: []string{
				"/src/lefthook.yaml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
			},
		},
		{
			name: "with keep config arg",
			args: UninstallArgs{KeepConfiguration: true},
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":  "not a lefthook hook",
				"/src/.git/hooks/post-commit": "# LEFTHOOK",
				"/src/lefthook.yml":           "# empty",
			},
			wantExist: []string{
				"/src/.git/hooks/pre-commit",
				"/src/lefthook.yml",
			},
			wantNotExist: []string{
				"/src/.git/hooks/post-commit",
			},
		},
		{
			name: "with .old files",
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":      "not a lefthook hook",
				"/src/.git/hooks/post-commit":     "LEFTHOOK file",
				"/src/.git/hooks/post-commit.old": "not a lefthook hook",
				"/src/lefthook.yml":               "# empty",
			},
			wantExist: []string{
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
			},
			wantNotExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/post-commit.old",
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			// Prepare files that should exist
			for file, content := range tt.existingFiles {
				if err := fs.MkdirAll(filepath.Base(file), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, file, []byte(content), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Do uninstall
			err := lefthook.Uninstall(&tt.args)
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
