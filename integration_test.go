//go:build integration

package main_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestLefthook(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("tests", "integration"),
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars, fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("GOCOVERDIR")))
			return nil
		},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"setup_git": func(ts *testscript.TestScript, _neg bool, _args []string) {
				if err := ts.Exec("git", "init"); err != nil {
					ts.Fatalf("unexpected error: %s", err)
				}
				if err := ts.Exec("git", "config", "user.email", "test@lefthook.dev"); err != nil {
					ts.Fatalf("unexpected error: %s", err)
				}
				if err := ts.Exec("git", "config", "user.name", "Jason Bourne"); err != nil {
					ts.Fatalf("unexpected error: %s", err)
				}
			},
			"git": func(ts *testscript.TestScript, neg bool, args []string) {
				err := ts.Exec("git", args...)
				if neg && err == nil {
					ts.Fatalf("expected error, got success")
				} else if !neg && err != nil {
					ts.Fatalf("expected success, got: %s", err)
				}
			},
			"lefthook": func(ts *testscript.TestScript, neg bool, args []string) {
				err := ts.Exec("lefthook", args...)
				if neg && err == nil {
					ts.Fatalf("expected error, got success")
				} else if !neg && err != nil {
					ts.Fatalf("expected success, got: %s", err)
				}
			},
		},
	})
}
