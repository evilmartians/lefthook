package runner

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
)

type TestExecutor struct{}

func (e TestExecutor) Execute(root string, args []string) (out *bytes.Buffer, err error) {
	out = bytes.NewBuffer(make([]byte, 0))

	if args[0] == "success" {
		err = nil
	} else {
		err = errors.New(args[0])
	}

	return
}

func TestRunAll(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	gitPath := filepath.Join(root, ".git")
	repo := &git.Repository{
		HooksPath: filepath.Join(gitPath, "hooks"),
		RootPath:  root,
		GitPath:   gitPath,
		InfoPath:  filepath.Join(gitPath, "info"),
	}

	for i, tt := range [...]struct {
		name          string
		args          []string
		scriptDirs    []string
		existingFiles []string
		hook          *config.Hook
		success, fail []Result
	}{
		{
			name: "empty hook",
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts:  map[string]*config.Script{},
				Piped:    true,
			},
		},
		{
			name: "with simple command",
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{{Name: "test", Status: StatusOk}},
		},
		{
			name: "with multiple commands ran in parallel",
			hook: &config.Hook{
				Parallel: true,
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
					"type-check": {
						Run: "fail",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{
				{Name: "test", Status: StatusOk},
				{Name: "lint", Status: StatusOk},
			},
			fail: []Result{{Name: "type-check", Status: StatusErr}},
		},
		{
			name: "with exclude tags",
			hook: &config.Hook{
				ExcludeTags: []string{"tests"},
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Tags: []string{"tests"},
					},
					"lint": {
						Run:  "success",
						Tags: []string{"linters"},
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{{Name: "lint", Status: StatusOk}},
		},
		{
			name: "with skip boolean option",
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Skip: true,
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{{Name: "lint", Status: StatusOk}},
		},
		{
			name: "with skip merge",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Skip: "merge",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{{Name: "lint", Status: StatusOk}},
		},
		{
			name: "with skip rebase and merge in an array",
			existingFiles: []string{
				filepath.Join(gitPath, "rebase-merge"),
				filepath.Join(gitPath, "rebase-apply"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Skip: []interface{}{"merge", "rebase"},
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{{Name: "lint", Status: StatusOk}},
		},
		{
			name: "with fail test",
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:      "fail",
						FailText: "try 'success'",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			fail: []Result{{Name: "test", Status: StatusErr, Text: "try 'success'"}},
		},
		{
			name: "with simple scripts",
			scriptDirs: []string{
				filepath.Join(root, config.DefaultSourceDir),
			},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "failing.js"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts: map[string]*config.Script{
					"script.sh": {
						Runner: "success",
					},
					"failing.js": {
						Runner:   "fail",
						FailText: "install node",
					},
				},
			},
			success: []Result{{Name: "script.sh", Status: StatusOk}},
			fail:    []Result{{Name: "failing.js", Status: StatusErr, Text: "install node"}},
		},
	} {
		fs := afero.NewMemMapFs()
		repo.Fs = fs
		resultChan := make(chan Result, len(tt.hook.Commands)+len(tt.hook.Scripts))
		executor := TestExecutor{}
		runner := &Runner{
			fs:         fs,
			repo:       repo,
			hook:       tt.hook,
			args:       tt.args,
			resultChan: resultChan,
			exec:       executor,
		}

		for _, file := range tt.existingFiles {
			if err := fs.MkdirAll(filepath.Base(file), 0o755); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if err := afero.WriteFile(fs, file, []byte{}, 0o755); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			runner.RunAll(tt.scriptDirs)
			close(resultChan)

			var success, fail []Result
			for res := range resultChan {
				if res.Status == StatusOk {
					success = append(success, res)
				} else {
					fail = append(fail, res)
				}
			}

			if !resultsMatch(success, tt.success) {
				t.Errorf("success results are not matching")
			}

			if !resultsMatch(fail, tt.fail) {
				t.Errorf("fail results are not matching")
			}
		})
	}
}

func resultsMatch(a, b []Result) bool {
	if len(a) != len(b) {
		return false
	}

	matches := make(map[string]struct{})

	for _, item := range a {
		str := fmt.Sprintf("%s_%d_%s", item.Name, item.Status, item.Text)
		matches[str] = struct{}{}
	}

	for _, item := range b {
		str := fmt.Sprintf("%s_%d_%s", item.Name, item.Status, item.Text)
		if _, ok := matches[str]; !ok {
			return false
		}
	}

	return true
}
