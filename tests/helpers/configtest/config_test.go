package configtest

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/config"
)

func TestParseHook(t *testing.T) {
	for i, tt := range [...]struct {
		raw  string
		hook *config.Hook
	}{
		{
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
			hook: &config.Hook{
				Parallel:    true,
				ExcludeTags: []string{"tag1", "tag2"},
				Jobs: []*config.Job{
					{
						Run: "echo",
					},
				},
				Commands: map[string]*config.Command{
					"simple": {
						Run: "echo",
					},
				},
				Scripts: map[string]*config.Script{
					"dummy.sh": {
						Runner: "bash",
					},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			parsed := ParseHook(tt.raw)
			assert.New(t).Equal(tt.hook, parsed)
		})
	}
}
