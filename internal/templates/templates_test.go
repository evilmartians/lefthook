package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHook_QuotesPathsWithSpaces(t *testing.T) {
	binPath := filepath.Join("/tmp", "my project", "bin", "lefthook")
	out := string(Hook("pre-commit", Args{LefthookPath: binPath}))

	slashPath := filepath.ToSlash(binPath)
	if !strings.Contains(out, `"`+slashPath+`" "$@"`) {
		t.Fatalf("expected quoted LefthookPath invocation, got:\n%s", out)
	}
	if strings.Contains(out, "elif "+slashPath+" -h") {
		t.Fatalf("unquoted path with space still present:\n%s", out)
	}
}

func TestHook_QuotesRcPathWithSpaces(t *testing.T) {
	rc := filepath.Join("/tmp", "my project", ".lefthookrc")
	out := string(Hook("pre-commit", Args{Rc: rc}))

	slashRc := filepath.ToSlash(rc)
	want := `[ -f "` + slashRc + `" ] && . "` + slashRc + `"`
	if !strings.Contains(out, want) {
		t.Fatalf("expected quoted rc path %q in:\n%s", want, out)
	}
}

func TestShellDoubleQuoteEscapesEmbeddedQuotes(t *testing.T) {
	got := shellDoubleQuote(`a"b\c$`)
	if got != `a\"b\\c\$` {
		t.Fatalf("unexpected escape: %q", got)
	}
}

func TestHook_QuotesExecutablePathWithSpaces(t *testing.T) {
	t.Chdir(t.TempDir())
	binDir := filepath.Join(t.TempDir(), "space dir")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(binDir, "lefthook")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Simulate install-time executable path baking.
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	out := string(Hook("pre-commit", Args{}))
	if !strings.Contains(out, `"`+filepath.ToSlash(bin)+`" "$@"`) {
		t.Fatalf("expected quoted executable path in hook:\n%s", out)
	}
}
