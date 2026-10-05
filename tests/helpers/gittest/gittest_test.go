package gittest_test

import (
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

func TestRepositoryBuilder_Build(t *testing.T) {
	fs := afero.NewMemMapFs()
	cmd := cmdtest.NewSpyCmd(nil)

	repo := gittest.NewRepositoryBuilder().Root("root").Fs(fs).Cmd(cmd).Build()

	want := &wrapper.Paths{
		Root:  "root",
		Git:   filepath.Join("root", ".git"),
		Info:  filepath.Join("root", ".git", "info"),
		Hooks: filepath.Join("root", ".git", "hooks"),
	}
	if !cmp.Equal(repo.Paths, want) {
		t.Errorf("repo.Paths = %v, want %v", repo.Paths, want)
	}

	if repo.Fs != fs {
		t.Errorf("repo.Fs = %v, want %v", repo.Fs, fs)
	}

	if len(cmd.Commands) != 0 {
		t.Errorf("cmd.Commands = %v, want empty", cmd.Commands)
	}
}

func TestRepositoryBuilder_Build_nextCalls(t *testing.T) {
	cmd := cmdtest.NewSpyCmd(func(_ string, _ string, out io.Writer) error {
		_, err := out.Write([]byte(strings.Join([]string{"new", "new/hooks", "new/info", "new/.git"}, "\n")))
		return err
	})

	repo := gittest.NewRepositoryBuilder().Root("root").Fs(afero.NewMemMapFs()).Cmd(cmd).Build()

	if err := repo.ResetPaths(); err != nil {
		t.Fatalf("repo.ResetPaths() error = %v, want nil", err)
	}

	if len(cmd.Commands) != 1 || !strings.HasPrefix(cmd.Commands[0], "git rev-parse") {
		t.Errorf("cmd.Commands = %v, want one git rev-parse call", cmd.Commands)
	}

	if repo.Paths.Root != "new" {
		t.Errorf("repo.Paths.Root = %v, want %v", repo.Paths.Root, "new")
	}
}

func TestGitPath(t *testing.T) {
	if result, want := gittest.GitPath("root"), filepath.Join("root", ".git"); result != want {
		t.Errorf("gittest.GitPath() = %v, want %v", result, want)
	}
}
