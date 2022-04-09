package lefthook

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/git"
)

func TestRun(t *testing.T) {
	repo := &git.Repository{
		HooksPath: "/src/.git/hooks",
		RootPath:  "/src/",
	}

	for i, tt := range [...]struct {
		name    string
		hook    string
		gitArgs []string
		envs    map[string]string
		error   bool
	}{
		{
			name: "Skip case",
			hook: "any-hook",
			envs: map[string]string{
				"LEFTHOOK": "0",
			},
			error: false,
		},
		{
			name: "Skip case",
			hook: "any-hook",
			envs: map[string]string{
				"LEFTHOOK": "false",
			},
			error: false,
		},
		// TODO: Add more testcases
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			fs := afero.NewMemMapFs()
			lefthook := &Lefthook{Options: &Options{Fs: fs}, repo: repo}

			for env, value := range tt.envs {
				t.Setenv(env, value)
			}

			err := lefthook.Run(tt.hook, tt.gitArgs)
			if err != nil {
				if !tt.error {
					t.Errorf("unexpected error :%s", err)
				}
			} else {
				if tt.error {
					t.Errorf("expected an error")
				}
			}
		})
	}
}
