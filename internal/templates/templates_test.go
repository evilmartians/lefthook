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

func TestHook_PreservesRcShellExpansion(t *testing.T) {
	rc := `${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc`
	out := string(Hook("pre-commit", Args{Rc: rc}))

	if strings.Contains(out, `\$HOME`) || strings.Contains(out, `\$XDG_CONFIG_HOME`) {
		t.Fatalf("rc path should preserve shell expansion tokens:\n%s", out)
	}
	if !strings.Contains(out, `[ -f ${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc ]`) {
		t.Fatalf("expected unquoted rc path with shell expansion in:\n%s", out)
	}
}

func TestHook_PreservesConfiguredShellCommand(t *testing.T) {
	out := string(Hook("pre-commit", Args{LefthookPath: "bundle exec lefthook"}))

	if strings.Contains(out, `"bundle exec lefthook" "$@"`) {
		t.Fatalf("configured shell command must not be quoted as one word:\n%s", out)
	}
	if !strings.Contains(out, "bundle exec lefthook \"$@\"") {
		t.Fatalf("expected unquoted shell command invocation in:\n%s", out)
	}
	if !strings.Contains(out, `elif test -n "bundle exec lefthook"`) {
		t.Fatalf("expected quoted test -n value for shell command in:\n%s", out)
	}
}

func TestShellEscapeDoubleQuoteEscapesEmbeddedQuotes(t *testing.T) {
	got := shellEscapeDoubleQuote(`a"b\c$`)
	if got != `a\"b\\c$` {
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
