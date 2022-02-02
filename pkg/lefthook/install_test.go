package lefthook

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
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

func TestLefthookInstall(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	repo := repoTest{
		hooksPath: "/.git/hooks",
		rootPath:  "/",
		gitPath:   "/",
	}
	hooksPath, _ := repo.HooksPath()

	for n, tt := range [...]struct {
		name, config string
		lefthook     Lefthook
		args         InstallArgs
		createdFiles []string
		wantExist    []string
	}{
		{
			name:     "default",
			lefthook: Lefthook{fs: fs, repo: repo, opts: &Options{}},
			args:     InstallArgs{},
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
			wantExist: []string{"pre-commit", "post-commit"},
		},
		{
			name:     "with existing hooks",
			lefthook: Lefthook{fs: fs, repo: repo, opts: &Options{}},
			args:     InstallArgs{},
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
			createdFiles: []string{filepath.Join(hooksPath, "pre-commit")},
			wantExist:    []string{"pre-commit", "post-commit", "pre-commit.old"},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", n, tt.name), func(t *testing.T) {
			if err := fs.WriteFile("/lefthook.yml", []byte(tt.config), 0644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for _, file := range tt.createdFiles {
				if err := fs.MkdirAll(filepath.Base(file), 0664); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				f, err := fs.Create(file)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				f.Close()
			}

			if err := tt.lefthook.Install(&tt.args); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			for _, file := range tt.wantExist {
				ok, err := fs.Exists(filepath.Join(hooksPath, file))
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if !ok {
					t.Errorf("expected %s to exist", file)
				}
			}
		})
	}
}
