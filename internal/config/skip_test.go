package config

import (
	"testing"

	"github.com/evilmartians/lefthook/internal/git"
)

type mockExecutor struct{}

func (mc mockExecutor) Cmd(cmd string) bool {
	return cmd == "success"
}

func TestDoSkip(t *testing.T) {
	skipChecker := NewSkipChecker(mockExecutor{})

	for _, tt := range [...]struct {
		name       string
		state      git.State
		skip, only interface{}
		skipped    bool
	}{
		{
			name:    "when true",
			state:   git.State{},
			skip:    true,
			skipped: true,
		},
		{
			name:    "when false",
			state:   git.State{},
			skip:    false,
			skipped: false,
		},
		{
			name:    "when merge",
			state:   git.State{Step: "merge"},
			skip:    "merge",
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   git.State{Step: "rebase"},
			skip:    "merge",
			skipped: false,
		},
		{
			name:    "when rebase",
			state:   git.State{Step: "rebase"},
			skip:    []interface{}{"rebase"},
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   git.State{Step: "rebase"},
			skip:    []interface{}{"merge"},
			skipped: false,
		},
		{
			name:    "when branch",
			state:   git.State{Branch: "feat/skipme"},
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			skipped: true,
		},
		{
			name:    "when branch doesn't match",
			state:   git.State{Branch: "feat/important"},
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			skipped: false,
		},
		{
			name:    "when branch glob",
			state:   git.State{Branch: "feat/important"},
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: true,
		},
		{
			name:    "when branch glob doesn't match",
			state:   git.State{Branch: "feat"},
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only specified",
			state:   git.State{Branch: "feat"},
			only:    []interface{}{map[string]interface{}{"ref": "feat"}},
			skipped: false,
		},
		{
			name:    "when only branch doesn't match",
			state:   git.State{Branch: "dev"},
			only:    []interface{}{map[string]interface{}{"ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only branch with glob",
			state:   git.State{Branch: "feat/important"},
			only:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only merge",
			state:   git.State{Step: "merge"},
			only:    []interface{}{"merge"},
			skipped: false,
		},
		{
			name:    "when only and skip",
			state:   git.State{Step: "rebase"},
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			only:    "rebase",
			skipped: false,
		},
		{
			name:    "when only and skip applies skip",
			state:   git.State{Step: "rebase"},
			skip:    []interface{}{"rebase"},
			only:    "rebase",
			skipped: true,
		},
		{
			name:    "when skip with run command",
			state:   git.State{},
			skip:    []interface{}{map[string]interface{}{"run": "success"}},
			skipped: true,
		},
		{
			name:    "when skip with multi-run command",
			state:   git.State{Branch: "feat"},
			skip:    []interface{}{map[string]interface{}{"run": "success", "ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only with run command",
			state:   git.State{},
			only:    []interface{}{map[string]interface{}{"run": "fail"}},
			skipped: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if skipChecker.DoSkip(tt.state, tt.skip, tt.only) != tt.skipped {
				t.Errorf("Expected: %v, Was %v", tt.skipped, !tt.skipped)
			}
		})
	}
}
