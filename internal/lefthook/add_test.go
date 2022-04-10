package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestLefthookAdd(t *testing.T) {
	repo := &git.Repository{
		HooksPath: hooksPath,
		RootPath:  root,
	}

	for n, tt := range [...]struct {
		name                    string
		args                    *AddArgs
		existingHooks           map[string]string
		config                  string
		wantExist, wantNotExist []string
		wantError               bool
	}{
		{
			name: "default empty repository",
			args: &AddArgs{Hook: "pre-commit"},
			wantExist: []string{
				hookPath("pre-commit"),
			},
			wantNotExist: []string{
				filepath.Join(root, ".lefthook"),
				filepath.Join(root, ".lefthook-local"),
			},
		},
		{
			name:      "unavailable hook",
			args:      &AddArgs{Hook: "super-star"},
			wantError: true,
			wantNotExist: []string{
				hookPath("super-star"),
				filepath.Join(root, ".lefthook"),
				filepath.Join(root, ".lefthook-local"),
			},
		},
		{
			name: "with create dirs arg",
			args: &AddArgs{Hook: "post-commit", CreateDirs: true},
			wantExist: []string{
				hookPath("post-commit"),
				filepath.Join(root, ".lefthook"),
				filepath.Join(root, ".lefthook-local"),
			},
		},
		{
			name: "with configured source dirs",
			args: &AddArgs{Hook: "post-commit", CreateDirs: true},
			config: `
source_dir: .source_dir
source_dir_local: .source_dir_local
`,
			wantExist: []string{
				hookPath("post-commit"),
				filepath.Join(root, ".source_dir", "post-commit"),
				filepath.Join(root, ".source_dir_local", "post-commit"),
			},
		},
		{
			name: "with existing hook",
			args: &AddArgs{Hook: "post-commit"},
			existingHooks: map[string]string{
				"post-commit": "custom script",
			},
			wantExist: []string{
				hookPath("post-commit"),
				hookPath("post-commit.old"),
			},
		},
		{
			name: "with existing lefthook hook",
			args: &AddArgs{Hook: "post-commit"},
			existingHooks: map[string]string{
				"post-commit": "LEFTHOOK file",
			},
			wantExist: []string{
				hookPath("post-commit"),
			},
			wantNotExist: []string{
				hookPath("post-commit.old"),
			},
		},
		{
			name: "with existing .old hook",
			args: &AddArgs{Hook: "post-commit"},
			existingHooks: map[string]string{
				"post-commit":     "custom hook",
				"post-commit.old": "custom old hook",
			},
			wantError: true,
			wantExist: []string{
				hookPath("post-commit"),
				hookPath("post-commit.old"),
			},
		},
		{
			name: "with existing .old hook, forced",
			args: &AddArgs{Hook: "post-commit", Force: true},
			existingHooks: map[string]string{
				"post-commit":     "custom hook",
				"post-commit.old": "custom old hook",
			},
			wantExist: []string{
				hookPath("post-commit"),
				hookPath("post-commit.old"),
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			if len(tt.config) > 0 {
				err := afero.WriteFile(fs, configPath, []byte(tt.config), 0o644)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			for hook, content := range tt.existingHooks {
				path := hookPath(hook)
				if err := fs.MkdirAll(filepath.Base(path), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			err := lefthook.Add(tt.args)
			if tt.wantError && err == nil {
				t.Errorf("expected an error")
			} else if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %s", err)
			}

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
