//go:build integration

package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestLefthook(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("tests", "integration"),
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars, fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("GOCOVERDIR")))
			return nil
		},
	})
}

func TestManualPrePushOpenStdin(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte("pre-push:\n  jobs:\n    - run: echo MANUAL {push_files}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "you@example.com"},
		{"config", "user.name", "Your Name"},
		{"add", "lefthook.yml"},
		{"commit", "-m", "initial"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "lefthook", "run", "pre-push")
	cmd.Dir = dir
	cmd.Stdin = reader
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("manual pre-push with open stdin: %s: %v", out, err)
	}
	if !strings.Contains(string(out), "MANUAL lefthook.yml") {
		t.Errorf("manual pre-push output = %q, want HEAD fallback", out)
	}
}
