package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/tests/helpers/configtest"
)

func TestScope_extend(t *testing.T) {
	for i, tt := range [...]struct {
		initial *scope
		job     *config.Job
		result  *scope
	}{
		{
			initial: newScope(&config.Hook{}, Options{}),
			job: configtest.ParseJob(`
        run: echo
        glob:
          - "*.js"
          - "*.jsx"
        env:
          VERSION: 1
          UI_ENABLE: false
          SERVICE_TOKEN: "secret"
        files: ls -A
        root: subdir/
      `),
			result: &scope{
				excludeFiles: []interface{}{},
				glob:         []string{"*.js", "*.jsx"},
				env: map[string]string{
					"VERSION":       "1",
					"UI_ENABLE":     "false",
					"SERVICE_TOKEN": "secret",
				},
				filesCmd: "ls -A",
				root:     "subdir/",
			},
		},
		{
			initial: &scope{
				fileTypes: []string{
					"text",
					"not executable",
				},
			},
			job: configtest.ParseJob(`
         file_types:
           - not symlink
      `),
			result: &scope{
				fileTypes: []string{
					"text",
					"not executable",
					"not symlink",
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := tt.initial.extend(tt.job)
			assert.Equal(t, tt.result, result)
		})
	}
}
