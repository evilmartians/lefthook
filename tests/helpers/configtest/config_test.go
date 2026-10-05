package configtest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/tests/helpers/configtest"
)

func TestParseHook(t *testing.T) {
	for name, tt := range map[string]struct {
		raw  string
		want *config.Hook
	}{
		"space-padding": {
			raw: `
        parallel: true
        exclude_tags:
          - tag1
          - tag2
        jobs:
          - run: echo
        commands:
          simple:
            run: echo
        scripts:
          "dummy.sh":
            runner: bash
      `,
			want: &config.Hook{
				Parallel:    true,
				ExcludeTags: []string{"tag1", "tag2"},
				Jobs:        []*config.Job{{Run: "echo"}},
				Commands:    map[string]*config.Command{"simple": {Run: "echo"}},
				Scripts:     map[string]*config.Script{"dummy.sh": {Runner: "bash"}},
			},
		},
		"no-padding": {
			raw:  "piped: true\njobs:\n  - run: echo\n",
			want: &config.Hook{Piped: true, Jobs: []*config.Job{{Run: "echo"}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := configtest.ParseHook(tt.raw)

			if !cmp.Equal(result, tt.want) {
				t.Errorf("configtest.ParseHook() = %v, want %v\n%s", result, tt.want, cmp.Diff(tt.want, result))
			}
		})
	}
}

func TestParseJob(t *testing.T) {
	for name, tt := range map[string]struct {
		raw  string
		want *config.Job
	}{
		"space-padding": {
			raw: `
        name: test
        run: echo
        glob:
          - "*.sh"
          - "*.md"
        exclude:
          - "install.sh"
          - "README.md"
        root: docs/
        use_stdin: true
        stage_fixed: true
      `,
			want: &config.Job{
				Name:       "test",
				Run:        "echo",
				Glob:       []string{"*.sh", "*.md"},
				Exclude:    []string{"install.sh", "README.md"},
				Root:       "docs/",
				UseStdin:   true,
				StageFixed: true,
			},
		},
		"tab-padding": {
			raw:  "\n\t\tname: test\n\t\trun: echo\n\t",
			want: &config.Job{Name: "test", Run: "echo"},
		},
		"no-padding": {
			raw:  "name: test\nrun: echo",
			want: &config.Job{Name: "test", Run: "echo"},
		},
		"leading-newlines": {
			raw:  "\n\n\n  name: test\n  run: echo\n\n",
			want: &config.Job{Name: "test", Run: "echo"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := configtest.ParseJob(tt.raw)

			if !cmp.Equal(result, tt.want) {
				t.Errorf("configtest.ParseJob() = %v, want %v\n%s", result, tt.want, cmp.Diff(tt.want, result))
			}
		})
	}
}

func TestParseHook_invalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("configtest.ParseHook() did not panic")
		}
	}()

	configtest.ParseHook("jobs: [")
}

func TestParseJob_invalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("configtest.ParseJob() did not panic")
		}
	}()

	configtest.ParseJob("glob: [")
}
