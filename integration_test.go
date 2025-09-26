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
			schemapath, err := filepath.Abs("schema.json")
			if err != nil {
				return err
			}

			env.Vars = append(env.Vars, fmt.Sprintf("GOCOVERDIR=%s", os.Getenv("GOCOVERDIR")))
			env.Vars = append(env.Vars, fmt.Sprintf("SCHEMAPATH=%s", schemapath))
			return nil
		},
	})
}
