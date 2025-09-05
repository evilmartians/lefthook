//go:build integrity
// +build integrity

package main_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestLefthookIntegrity(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("tests", "integrity"),
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars, fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("GOCOVERDIR")))
			return nil
		},
	})
}
