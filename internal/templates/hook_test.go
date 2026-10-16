package templates

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// renderHook renders the pre-push hook with no binary path, so every
// invocation exercises the not-found branch.
func renderHook(t *testing.T, assert bool) string {
	t.Helper()
	return string(Hook("pre-push", Args{
		LefthookPath:            "",
		AssertLefthookInstalled: assert,
		Roots:                   []string{"."},
	}))
}

// TestHookFailClosedChecksEveryConfigCandidate asserts that the generated
// not-found branch iterates over every main/local config name and extension
// the loader itself searches (internal/config/loader.go). The template builds
// the candidates as a names x extensions loop, so the assertion pins the loop
// element lists: with both lists complete, the cross product covers all 30
// candidates and cannot silently drift from the supported set.
func TestHookFailClosedChecksEveryConfigCandidate(t *testing.T) {
	script := strings.ReplaceAll(renderHook(t, false), "\r", "")
	var baseList, extList string
	lines := strings.Split(script, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "for base in") {
			// Consume backslash continuations: the name list wraps.
			for baseList = line; strings.HasSuffix(strings.TrimSpace(baseList), "\\") && i+1 < len(lines); {
				i++
				baseList += " " + lines[i]
			}
		}
		if strings.Contains(line, "for ext in") {
			extList = line
		}
	}
	if baseList == "" || extList == "" {
		t.Fatal("generated hook has no config-name/extension scan loop")
	}
	for _, name := range []string{
		"lefthook", ".lefthook", ".config/lefthook",
		"lefthook-local", ".lefthook-local", ".config/lefthook-local",
	} {
		if !strings.Contains(baseList, name) {
			t.Errorf("config scan loop misses config name %q", name)
		}
	}
	for _, ext := range []string{".yml", ".yaml", ".json", ".jsonc", ".toml"} {
		if !strings.Contains(extList, ext) {
			t.Errorf("config scan loop misses extension %q", ext)
		}
	}
	if !strings.Contains(script, "LEFTHOOK_CONFIG") {
		t.Error("generated hook does not honor the LEFTHOOK_CONFIG override")
	}
}

// writeHookScript renders the hook to an executable file and returns its
// path. The generated shim is a POSIX-sh script, so execution tests are
// skipped on Windows; the candidate-coverage test above still runs there.
func writeHookScript(t *testing.T, assert bool) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("executes the generated POSIX-sh hook; skipped on windows")
	}
	path := filepath.Join(t.TempDir(), "pre-push")
	if err := os.WriteFile(path, []byte(renderHook(t, assert)), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// runHook executes the shim the way git would (argv: <url> <ref>) in the
// given working directory, with a minimal environment so the lefthook binary
// is never found.
func runHook(t *testing.T, script, dir string, extraEnv ...string) int {
	t.Helper()
	cmd := exec.Command(script, "origin", "fake-ref")
	cmd.Dir = dir
	cmd.Env = append([]string{"PATH=/usr/bin:/bin", "HOME=" + os.TempDir()}, extraEnv...)
	err := cmd.Run()
	if err == nil {
		return 0
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("hook run failed unexpectedly: %v", err)
	}
	return exitErr.ExitCode()
}

func TestHookFailsClosedWhenConfigPresent(t *testing.T) {
	script := writeHookScript(t, false)
	repo := t.TempDir()
	// The TOML variant on purpose: it is the case most likely to be missed.
	if err := os.WriteFile(filepath.Join(repo, "lefthook.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runHook(t, script, repo); code != 1 {
		t.Errorf("exit = %d, want 1 (fail closed when a config exists)", code)
	}
}

func TestHookStaysNoOpWithoutConfig(t *testing.T) {
	script := writeHookScript(t, false)
	if code := runHook(t, script, t.TempDir()); code != 0 {
		t.Errorf("exit = %d, want 0 (no-op without a lefthook config)", code)
	}
}

func TestHookHonorsLefthookDisabled(t *testing.T) {
	script := writeHookScript(t, false)
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "lefthook.yml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runHook(t, script, repo, "LEFTHOOK=0"); code != 0 {
		t.Errorf("exit = %d, want 0 (LEFTHOOK=0 escape hatch)", code)
	}
}

func TestHookOverrideConfigFailsClosed(t *testing.T) {
	script := writeHookScript(t, false)
	repo := t.TempDir()

	// Absolute override (any OS form): used as-is, like the loader does.
	abs := filepath.Join(t.TempDir(), "override.toml")
	if err := os.WriteFile(abs, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runHook(t, script, repo, "LEFTHOOK_CONFIG="+abs); code != 1 {
		t.Errorf("exit = %d, want 1 (absolute LEFTHOOK_CONFIG present)", code)
	}

	// Relative override: resolved against the repository root.
	if err := os.WriteFile(filepath.Join(repo, "rel.yml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runHook(t, script, repo, "LEFTHOOK_CONFIG=rel.yml"); code != 1 {
		t.Errorf("exit = %d, want 1 (relative LEFTHOOK_CONFIG present)", code)
	}
}

func TestHookAssertRenderingAlwaysFailsClosed(t *testing.T) {
	script := renderHook(t, true)
	if !strings.Contains(script, "exit 1") {
		t.Error("assert_lefthook_installed rendering must abort when the binary is missing")
	}
}
