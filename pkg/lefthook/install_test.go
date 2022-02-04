package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/pkg/git"
)

type repoTest struct {
	hooksPath, rootPath, gitPath string
}

func (r repoTest) HooksPath() (string, error) {
	return r.hooksPath, nil
}

func (r repoTest) RootPath() string {
	return r.rootPath
}

func (r repoTest) GitPath() string {
	return r.gitPath
}

func (r repoTest) OperationInProgress() bool {
	return false
}

func RepoTest() git.Repository {
	return repoTest{
		hooksPath: "/src/.git/hooks",
		rootPath:  "/src/",
		gitPath:   "/src/",
	}
}

func TestLefthookInstall(t *testing.T) {
	repo := RepoTest()

	for n, tt := range [...]struct {
		name, config            string
		args                    InstallArgs
		existingFiles           map[string]string
		wantExist, wantNotExist []string
		wantError               bool
	}{
		{
			name: "without a config file",
			wantExist: []string{
				"/src/lefthook.yml",
			},
		},
		{
			name: "simple default config",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/" + checksumHookFilename,
			},
		},
		{
			name: "with existing hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit": "",
			},
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/pre-commit.old",
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/" + checksumHookFilename,
			},
		},
		{
			name: "with existing lefthook hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit": "# lefthook_version: 8b2c9fc6b3391b3cf020b97ab7037c61",
			},
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/" + checksumHookFilename,
			},
			wantNotExist: []string{
				"/src/.git/hooks/pre-commit.old",
			},
		},
		{
			name: "with synchronized hooks",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/prepare-commit-msg": "# lefthook_version: 8b2c9fc6b3391b3cf020b97ab7037c61",
			},
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/" + checksumHookFilename,
			},
			wantNotExist: []string{
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
			},
		},
		{
			name: "with synchronized hooks forced",
			args: InstallArgs{Force: true},
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/prepare-commit-msg": "# lefthook_version: 8b2c9fc6b3391b3cf020b97ab7037c61",
			},
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/" + checksumHookFilename,
			},
		},
		{
			name: "with existing hook and .old file",
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":     "",
				"/src/.git/hooks/pre-commit.old": "",
			},
			wantError: true,
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/pre-commit.old",
			},
			wantNotExist: []string{
				"/src/.git/hooks/" + checksumHookFilename,
			},
		},
		{
			name: "with existing hook and .old file, but forced",
			args: InstallArgs{Force: true},
			config: `
pre-commit:
  commands:
    tests:
      run: yarn test

post-commit:
  commands:
    notify:
      run: echo 'Done!'
`,
			existingFiles: map[string]string{
				"/src/.git/hooks/pre-commit":     "",
				"/src/.git/hooks/pre-commit.old": "",
			},
			wantExist: []string{
				"/src/lefthook.yml",
				"/src/.git/hooks/pre-commit",
				"/src/.git/hooks/pre-commit.old",
				"/src/.git/hooks/post-commit",
				"/src/.git/hooks/" + checksumHookFilename,
			},
		},
	} {
		fs := afero.NewMemMapFs()
		lefthook := Lefthook{fs: fs, repo: repo, opts: &Options{}}

		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			// Create configuration file
			if len(tt.config) > 0 {
				if err := afero.WriteFile(fs, "/src/lefthook.yml", []byte(tt.config), 0644); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Create files that should exist
			for file, content := range tt.existingFiles {
				if err := fs.MkdirAll(filepath.Base(file), 0664); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if err := afero.WriteFile(fs, file, []byte(content), 0755); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			// Do install
			err := lefthook.Install(&tt.args)
			if tt.wantError && err == nil {
				t.Errorf("expected an error")
			} else if !tt.wantError && err != nil {
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
