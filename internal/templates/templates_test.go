package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHook_OmitsInstallTimeExecutablePath(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	out := string(Hook("pre-commit", Args{}))
	if strings.Contains(out, filepath.ToSlash(exe)) {
		t.Fatalf("hook shim must not bake install-time executable path %q into shared .git/hooks", exe)
	}
}

func TestHook_KeepsWorktreeRelativeNodeModulesFallback(t *testing.T) {
	out := string(Hook("pre-commit", Args{}))

	if !strings.Contains(out, `dir="$(git rev-parse --show-toplevel)"`) {
		t.Fatal("expected worktree-relative node_modules fallback")
	}
	if !strings.Contains(out, `node_modules/lefthook-${osArch}-${cpuArch}/bin/lefthook`) {
		t.Fatal("expected per-worktree lefthook npm binary fallback")
	}
}
