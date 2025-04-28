package config

import (
	"errors"
	"io"
	"testing"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

type mockCmd struct{}

func (mc mockCmd) WithoutEnvs(...string) system.Command {
	return mc
}

func (mc mockCmd) Run(cmd []string, _root string, _in io.Reader, _out io.Writer, _errOut io.Writer) error {
	if len(cmd) == 3 && cmd[2] == "success" {
		return nil
	} else {
		return errors.New("failure")
	}
}

func TestSkipChecker_Check(t *testing.T) {
	skipChecker := NewSkipChecker(mockCmd{})

	for _, tt := range [...]struct {
		name       string
		state      func() git.State
		skip, only interface{}
		skipped    bool
	}{
		{
			name:    "when true",
			state:   func() git.State { return git.State{} },
			skip:    true,
			skipped: true,
		},
		{
			name:    "when false",
			state:   func() git.State { return git.State{} },
			skip:    false,
			skipped: false,
		},
		{
			name:    "when merge",
			state:   func() git.State { return git.State{State: "merge"} },
			skip:    "merge",
			skipped: true,
		},
		{
			name:    "when merge-commit",
			state:   func() git.State { return git.State{State: "merge-commit"} },
			skip:    "merge-commit",
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   func() git.State { return git.State{State: "rebase"} },
			skip:    "merge",
			skipped: false,
		},
		{
			name:    "when rebase",
			state:   func() git.State { return git.State{State: "rebase"} },
			skip:    []interface{}{"rebase"},
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   func() git.State { return git.State{State: "rebase"} },
			skip:    []interface{}{"merge"},
			skipped: false,
		},
		{
			name:    "when branch",
			state:   func() git.State { return git.State{Branch: "feat/skipme"} },
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			skipped: true,
		},
		{
			name:    "when branch doesn't match",
			state:   func() git.State { return git.State{Branch: "feat/important"} },
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			skipped: false,
		},
		{
			name:    "when branch glob",
			state:   func() git.State { return git.State{Branch: "feat/important"} },
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: true,
		},
		{
			name:    "when branch glob doesn't match",
			state:   func() git.State { return git.State{Branch: "feat"} },
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only specified",
			state:   func() git.State { return git.State{Branch: "feat"} },
			only:    []interface{}{map[string]interface{}{"ref": "feat"}},
			skipped: false,
		},
		{
			name:    "when only branch doesn't match",
			state:   func() git.State { return git.State{Branch: "dev"} },
			only:    []interface{}{map[string]interface{}{"ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only branch with glob",
			state:   func() git.State { return git.State{Branch: "feat/important"} },
			only:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only merge",
			state:   func() git.State { return git.State{State: "merge"} },
			only:    []interface{}{"merge"},
			skipped: false,
		},
		{
			name:    "when only and skip",
			state:   func() git.State { return git.State{State: "rebase"} },
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			only:    "rebase",
			skipped: false,
		},
		{
			name:    "when only and skip applies skip",
			state:   func() git.State { return git.State{State: "rebase"} },
			skip:    []interface{}{"rebase"},
			only:    "rebase",
			skipped: true,
		},
		{
			name:    "when skip with run command",
			state:   func() git.State { return git.State{} },
			skip:    []interface{}{map[string]interface{}{"run": "success"}},
			skipped: true,
		},
		{
			name:    "when skip with multi-run command",
			state:   func() git.State { return git.State{Branch: "feat"} },
			skip:    []interface{}{map[string]interface{}{"run": "success", "ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only with run command",
			state:   func() git.State { return git.State{} },
			only:    []interface{}{map[string]interface{}{"run": "fail"}},
			skipped: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if skipChecker.Check(tt.state, tt.skip, tt.only) != tt.skipped {
				t.Errorf("Expected: %v, Was %v", tt.skipped, !tt.skipped)
			}
		})
	}
}
