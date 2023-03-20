package config

import (
	"testing"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestIsSkip(t *testing.T) {
	for _, tt := range [...]struct {
		name    string
		skip    interface{}
		state   git.State
		skipped bool
	}{
		{
			name:    "when true",
			skip:    true,
			state:   git.State{},
			skipped: true,
		},
		{
			name:    "when false",
			skip:    false,
			state:   git.State{},
			skipped: false,
		},
		{
			name:    "when merge",
			skip:    "merge",
			state:   git.State{Step: "merge"},
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			skip:    "merge",
			state:   git.State{Step: "rebase"},
			skipped: false,
		},
		{
			name:    "when rebase",
			skip:    []interface{}{"rebase"},
			state:   git.State{Step: "rebase"},
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			skip:    []interface{}{"merge"},
			state:   git.State{Step: "rebase"},
			skipped: false,
		},
		{
			name:    "when branch",
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			state:   git.State{Branch: "feat/skipme"},
			skipped: true,
		},
		{
			name:    "when branch doesn't match",
			skip:    []interface{}{map[string]interface{}{"ref": "feat/skipme"}},
			state:   git.State{Branch: "feat/important"},
			skipped: false,
		},
		{
			name:    "when branch glob",
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			state:   git.State{Branch: "feat/important"},
			skipped: true,
		},
		{
			name:    "when branch glob doesn't match",
			skip:    []interface{}{map[string]interface{}{"ref": "feat/*"}},
			state:   git.State{Branch: "feat"},
			skipped: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if isSkip(tt.state, tt.skip) != tt.skipped {
				t.Errorf("Expected: %v, Was %v", tt.skipped, !tt.skipped)
			}
		})
	}
}
