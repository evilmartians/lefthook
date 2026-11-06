package command

import (
	"path/filepath"
	"testing"

	"al.essio.dev/pkg/shellescape"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestBuildScript(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	hookName := "echo-args"
	sourceDir := filepath.Join(root, ".lefthook")
	scriptPath := filepath.Join(sourceDir, hookName, "echo.sh")
	quotedScript := shellescape.Quote(scriptPath)

	for name, tt := range map[string]struct {
		runner  string
		gitArgs []string
		exec    string
	}{
		"no args": {
			runner: "sh",
			exec:   "sh " + quotedScript,
		},
		"simple args are passed as-is": {
			runner:  "sh",
			gitArgs: []string{"-Apply", "file.txt"},
			exec:    "sh " + quotedScript + " -Apply file.txt",
		},
		"args with spaces are kept intact": {
			runner:  "powershell",
			gitArgs: []string{"path with spaces.yml", "file (1).txt"},
			exec:    "powershell " + quotedScript + " 'path with spaces.yml' 'file (1).txt'",
		},
		"args with shell special characters are escaped": {
			runner:  "sh",
			gitArgs: []string{"a$b", "c`d", "e&f;g", "h\"i"},
			exec:    "sh " + quotedScript + " 'a$b' 'c`d' 'e&f;g' 'h\"i'",
		},
		"args with single quotes are escaped": {
			runner:  "sh",
			gitArgs: []string{"it's"},
			exec:    "sh " + quotedScript + ` 'it'"'"'s'`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fs := afero.NewMemMapFs()
			assert.NoError(fs.MkdirAll(filepath.Dir(scriptPath), 0o755))
			assert.NoError(afero.WriteFile(fs, scriptPath, []byte("#!/bin/sh\n"), 0o755))

			repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
			builder := NewBuilder(repo, loggertest.NewExecution(), BuilderOptions{
				HookName:   hookName,
				SourceDirs: []string{sourceDir},
				GitArgs:    tt.gitArgs,
			})

			execs, files, err := builder.BuildCommands(&JobParams{
				Name:   "echo.sh",
				Script: "echo.sh",
				Runner: tt.runner,
			})

			assert.NoError(err)
			assert.Empty(files)
			assert.Equal([]string{tt.exec}, execs)
		})
	}
}
