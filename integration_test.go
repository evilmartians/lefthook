//go:build integration
// +build integration

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
	})
}
