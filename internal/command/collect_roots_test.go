package command

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/evilmartians/lefthook/v2/internal/config"
)

func Test_collectRoots(t *testing.T) {
	for name, tt := range map[string]struct {
		hook *config.Hook
		want []string
	}{
		"no-roots": {
			hook: &config.Hook{
				Jobs: []*config.Job{{Run: "echo"}},
			},
			want: []string{},
		},
		"command-and-job-roots": {
			hook: &config.Hook{
				Commands: map[string]*config.Command{"lint": {Run: "echo", Root: "client/"}},
				Jobs:     []*config.Job{{Run: "echo", Root: "/server/"}},
			},
			want: []string{"client", "server"},
		},
		"group-root": {
			hook: &config.Hook{
				Jobs: []*config.Job{{
					Group: &config.Group{
						Root: "src/",
						Jobs: []*config.Job{
							{Run: "echo"},
							{Run: "echo", Root: "frontend/"},
						},
					},
				}},
			},
			want: []string{"frontend", "src"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := collectRoots(&config.Config{Hooks: map[string]*config.Hook{"pre-commit": tt.hook}})
			slices.Sort(got)

			if !cmp.Equal(got, tt.want) {
				t.Errorf("collectRoots() = %v, want %v", got, tt.want)
			}
		})
	}
}
