package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/tests/helpers/configtest"
)

func Test_newScope(t *testing.T) {
	t.Run("with excluded files in hook and opts", func(t *testing.T) {
		opts := Options{
			ExcludeFiles: []string{
				"file1.txt",
				"file2.txt",
			},
		}
		hook := &config.Hook{
			Exclude: []string{
				"file3.txt",
				"file4.txt",
				"file5.txt",
			},
		}

		scope := newScope(hook, opts)
		assert.Equal(t, scope.excludeFiles, []string{
			"file1.txt",
			"file2.txt",
			"file3.txt",
			"file4.txt",
			"file5.txt",
		})
		assert.NotEqual(t, scope.env, nil)
	})

	t.Run("without excluded files", func(t *testing.T) {
		opts := Options{}
		hook := &config.Hook{}

		scope := newScope(hook, opts)
		assert.Equal(t, scope.excludeFiles, []string{})
	})

	t.Run("without excluded files from hook only", func(t *testing.T) {
		opts := Options{}
		hook := &config.Hook{
			Exclude: []string{
				"file1.txt",
				"file2.txt",
			},
		}

		scope := newScope(hook, opts)
		assert.Equal(t, scope.excludeFiles, []string{
			"file1.txt",
			"file2.txt",
		})
	})
}

func TestScope_extend(t *testing.T) {
	for i, tt := range [...]struct {
		initial *scope
		job     *config.Job
		result  *scope
	}{
		{
			initial: &scope{},
			job: configtest.ParseJob(`
        run: echo
        glob:
          - "*.js"
          - "*.jsx"
        exclude:
          - "folder/*.sh"
      `),
			result: &scope{
				glob:         []string{"*.js", "*.jsx"},
				excludeFiles: []string{"folder/*.sh"},
			},
		},
		{
			initial: &scope{},
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
				glob: []string{"*.js", "*.jsx"},
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
