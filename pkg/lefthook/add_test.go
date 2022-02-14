package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/git"
)

func TestLefthookAdd(t *testing.T) {
	repo := &git.Repository{
		HooksPath: "/src/.git/hooks",
		RootPath:  "/src/",
		GitPath:   "/src/",
	}

	for n, tt := range [...]struct {
		name                    string
		args                    *AddArgs
		existingFiles           map[string]string
		wantExist, wantNotExist []string
		wantError               bool
	}{
		{
			name: "default empty repository",
			args: &AddArgs{Hook: "pre-commit"},
			wantExist: []string{
				"/src/.git/hooks/pre-commit",
			},
			wantNotExist: []string{
				"/src/.lefthook",
				"/src/.lefthook-local",
			},
		},
		{
			name:      "unavailable hook",
			args:      &AddArgs{Hook: "super-star"},
			wantError: true,
			wantNotExist: []string{
				"/src/.git/hooks/super-star",
				"/src/.lefthook",
				"/src/.lefthook-local",
			},
		},
		{
			name: "with create dirs arg",
			args: &AddArgs{Hook: "post-commit", CreateDirs: true},
			wantExist: []string{
				"/src/.git/hooks/post-commit",
				"/src/.lefthook/",
				"/src/.lefthook-local/",
			},
		},
		{
			name: "with configured source dirs",
			args: &AddArgs{Hook: "post-commit", CreateDirs: true},
			existingFiles: map[string]string{
				"/src/lefthook.yml": `
source_dir: .source_dir
source_dir_local: .source_dir_local
`,
			},
			wantExist: []string{
				"/src/.git/hooks/post-commit",
				"/src/.source_dir/post-commit/",
				"/src/.source_dir_local/post-commit/",
			},
		},
		{
			name: "with existing hook",
			args: &AddArgs{Hook: "post-commit"},
			existingFiles: map[string]string{
				"/src/.git/hooks/post-commit": "custom script",
			},
			wantExist: []string{
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/post-commit.old",
			},
		},
		{
			name: "with existing lefthook hook",
			args: &AddArgs{Hook: "post-commit"},
			existingFiles: map[string]string{
				"/src/.git/hooks/post-commit": "LEFTHOOK file",
			},
			wantExist: []string{
				"/src/.git/hooks/post-commit",
			},
			wantNotExist: []string{
				"/src/.git/hooks/post-commit.old",
			},
		},
		{
			name: "with existing .old hook",
			args: &AddArgs{Hook: "post-commit"},
			existingFiles: map[string]string{
				"/src/.git/hooks/post-commit":     "custom hook",
				"/src/.git/hooks/post-commit.old": "custom old hook",
			},
			wantError: true,
			wantExist: []string{
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/post-commit.old",
			},
		},
		{
			name: "with existing .old hook, forced",
			args: &AddArgs{Hook: "post-commit", Force: true},
			existingFiles: map[string]string{
				"/src/.git/hooks/post-commit":     "custom hook",
				"/src/.git/hooks/post-commit.old": "custom old hook",
			},
			wantExist: []string{
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/post-commit.old",
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			for file, content := range tt.existingFiles {
				if err := fs.MkdirAll(filepath.Base(file), 0o755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, file, []byte(content), 0o644); err != nil {
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
