package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/lefthook/run/exec"
)

type TestExecutor struct{}

func (e TestExecutor) Execute(_ctx context.Context, opts exec.Options, _out io.Writer) (err error) {
	if strings.HasPrefix(opts.Commands[0], "success") {
		err = nil
	} else {
		err = errors.New(opts.Commands[0])
	}

	return
}

func (e TestExecutor) RawExecute(_ctx context.Context, _command []string, _out io.Writer) error {
	return nil
}

type GitMock struct {
	mux      sync.Mutex
	commands []string
}

func (g *GitMock) SetRootPath(_root string) {}

func (g *GitMock) Cmd(cmd string) (string, error) {
	g.mux.Lock()
	g.commands = append(g.commands, cmd)
	g.mux.Unlock()

	return "", nil
}

func (g *GitMock) CmdArgs(args ...string) (string, error) {
	g.mux.Lock()
	g.commands = append(g.commands, strings.Join(args, " "))
	g.mux.Unlock()

	return "", nil
}

func (g *GitMock) CmdLines(cmd string) ([]string, error) {
	g.mux.Lock()
	g.commands = append(g.commands, cmd)
	g.mux.Unlock()

	if cmd == "git diff --name-only --cached --diff-filter=ACMR" ||
		cmd == "git diff --name-only HEAD @{push}" {
		root, _ := filepath.Abs("src")
		return []string{
			filepath.Join(root, "scripts", "script.sh"),
			filepath.Join(root, "README.md"),
		}, nil
	}

	return nil, nil
}

func (g *GitMock) RawCmd(cmd string) (string, error) {
	g.mux.Lock()
	g.commands = append(g.commands, cmd)
	g.mux.Unlock()

	return "", nil
}

func (g *GitMock) reset() {
	g.mux.Lock()
	g.commands = []string{}
	g.mux.Unlock()
}

func TestRunAll(t *testing.T) {
	root, err := filepath.Abs("src")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	gitExec := &GitMock{}
	gitPath := filepath.Join(root, ".git")
	repo := &git.Repository{
		Git:       gitExec,
		HooksPath: filepath.Join(gitPath, "hooks"),
		RootPath:  root,
		GitPath:   gitPath,
		InfoPath:  filepath.Join(gitPath, "info"),
	}

	for i, tt := range [...]struct {
		name, branch, hookName string
		args                   []string
		sourceDirs             []string
		existingFiles          []string
		hook                   *config.Hook
		success, fail          []Result
		gitCommands            []string
	}{
		{
			name:     "empty hook",
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts:  map[string]*config.Script{},
				Piped:    true,
			},
		},
		{
			name:     "with simple command",
			hookName: "post-commit",
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
			name:     "with simple command in follow mode",
			hookName: "post-commit",
			hook: &config.Hook{
				Follow: true,
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
			name:     "with multiple commands ran in parallel",
			hookName: "post-commit",
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
			name:     "with exclude tags",
			hookName: "post-commit",
			hook: &config.Hook{
				ExcludeTags: []string{"tests", "formatter"},
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Tags: []string{"tests"},
					},
					"formatter": {
						Run: "success",
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
			name:     "with skip boolean option",
			hookName: "post-commit",
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
			name:     "with skip merge",
			hookName: "post-commit",
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
			name:     "with only on merge",
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Only: "merge",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{
				{Name: "lint", Status: StatusOk},
				{Name: "test", Status: StatusOk},
			},
		},
		{
			name:     "with only on merge",
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"test": {
						Run:  "success",
						Only: "merge",
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
			name:     "with global skip merge",
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: &config.Hook{
				Skip: "merge",
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{},
		},
		{
			name:     "with global only on merge",
			hookName: "post-commit",
			hook: &config.Hook{
				Only: "merge",
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{},
		},
		{
			name:     "with global only on merge",
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: &config.Hook{
				Only: "merge",
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{
				{Name: "lint", Status: StatusOk},
				{Name: "test", Status: StatusOk},
			},
		},
		{
			name:     "with skip rebase and merge in an array",
			hookName: "post-commit",
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
			name:   "with global skip on ref",
			branch: "main",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Skip: []interface{}{"merge", map[string]interface{}{"ref": "main"}},
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{},
		},
		{
			name:   "with global only on ref",
			branch: "main",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Only: []interface{}{"merge", map[string]interface{}{"ref": "main"}},
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{
				{Name: "lint", Status: StatusOk},
				{Name: "test", Status: StatusOk},
			},
		},
		{
			name:   "with global only on ref",
			branch: "develop",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Only: []interface{}{"merge", map[string]interface{}{"ref": "main"}},
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{},
		},
		{
			name:   "with global skip on another ref",
			branch: "fix",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Skip: []interface{}{"merge", map[string]interface{}{"ref": "main"}},
				Commands: map[string]*config.Command{
					"test": {
						Run: "success",
					},
					"lint": {
						Run: "success",
					},
				},
				Scripts: map[string]*config.Script{},
			},
			success: []Result{
				{Name: "test", Status: StatusOk},
				{Name: "lint", Status: StatusOk},
			},
		},
		{
			name:     "with fail test",
			hookName: "post-commit",
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
			name:       "with simple scripts",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
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
		{
			name:       "with simple scripts and only option",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts: map[string]*config.Script{
					"script.sh": {
						Runner: "success",
						Only:   "merge",
					},
					"failing.js": {
						Only:     "merge",
						Runner:   "fail",
						FailText: "install node",
					},
				},
			},
			success: []Result{{Name: "script.sh", Status: StatusOk}},
			fail:    []Result{{Name: "failing.js", Status: StatusErr, Text: "install node"}},
		},
		{
			name:       "with simple scripts and only option",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts: map[string]*config.Script{
					"script.sh": {
						Only:   "merge",
						Runner: "success",
					},
					"failing.js": {
						Only:     "merge",
						Runner:   "fail",
						FailText: "install node",
					},
				},
			},
			success: []Result{},
			fail:    []Result{},
		},
		{
			name:       "with interactive and parallel",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Parallel: true,
				Commands: map[string]*config.Command{
					"ok": {
						Run:         "success",
						Interactive: true,
					},
					"fail": {
						Run: "fail",
					},
				},
				Scripts: map[string]*config.Script{
					"script.sh": {
						Runner:      "success",
						Interactive: true,
					},
					"failing.js": {
						Runner: "fail",
					},
				},
			},
			success: []Result{}, // script.sh and ok are skipped
			fail:    []Result{{Name: "failing.js", Status: StatusErr}, {Name: "fail", Status: StatusErr}},
		},
		{
			name:       "with stage_fixed in true",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "success.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run:        "success",
						StageFixed: true,
					},
					"fail": {
						Run:        "fail",
						StageFixed: true,
					},
				},
				Scripts: map[string]*config.Script{
					"success.sh": {
						Runner:     "success",
						StageFixed: true,
					},
					"failing.js": {
						Runner:     "fail",
						StageFixed: true,
					},
				},
			},
			success: []Result{{Name: "ok", Status: StatusOk}, {Name: "success.sh", Status: StatusOk}},
			fail:    []Result{{Name: "fail", Status: StatusErr}, {Name: "failing.js", Status: StatusErr}},
		},
		{
			name:       "pre-commit hook simple",
			hookName:   "pre-commit",
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "pre-commit", "success.sh"),
				filepath.Join(root, config.DefaultSourceDir, "pre-commit", "failing.js"),
				filepath.Join(root, "scripts", "script.sh"),
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run:        "success",
						StageFixed: true,
					},
					"fail": {
						Run:        "fail",
						StageFixed: true,
					},
				},
				Scripts: map[string]*config.Script{
					"success.sh": {
						Runner:     "success",
						StageFixed: true,
					},
					"failing.js": {
						Runner:     "fail",
						StageFixed: true,
					},
				},
			},
			success: []Result{{Name: "ok", Status: StatusOk}, {Name: "success.sh", Status: StatusOk}},
			fail:    []Result{{Name: "fail", Status: StatusErr}, {Name: "failing.js", Status: StatusErr}},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add .*script.sh.*README.md",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add .*script.sh.*README.md",
				"git apply -v --whitespace=nowarn --recount --unidiff-zero ",
				"git stash list",
			},
		},
		{
			name:     "pre-commit hook with implicit skip",
			hookName: "pre-commit",
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run:        "success",
						StageFixed: true,
						Glob:       "*.md",
					},
					"fail": {
						Run:        "fail",
						StageFixed: true,
						Glob:       "*.sh",
					},
				},
			},
			success: []Result{{Name: "ok", Status: StatusOk}},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add .*README.md",
				"git apply -v --whitespace=nowarn --recount --unidiff-zero ",
				"git stash list",
			},
		},
		{
			name:     "pre-commit hook with stage_fixed under root",
			hookName: "pre-commit",
			existingFiles: []string{
				filepath.Join(root, "scripts", "script.sh"),
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run:        "success",
						Root:       filepath.Join(root, "scripts"),
						StageFixed: true,
					},
				},
			},
			success: []Result{{Name: "ok", Status: StatusOk}},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add .*scripts.*script.sh",
				"git apply -v --whitespace=nowarn --recount --unidiff-zero ",
				"git stash list",
			},
		},
		{
			name:     "pre-push hook with implicit skip",
			hookName: "pre-push",
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run:        "success",
						StageFixed: true,
						Glob:       "*.md",
					},
					"fail": {
						Run:        "fail",
						StageFixed: true,
						Glob:       "*.sh",
					},
				},
			},
			success: []Result{{Name: "ok", Status: StatusOk}},
			gitCommands: []string{
				"git diff --name-only HEAD @{push}",
				"git diff --name-only HEAD @{push}",
			},
		},
	} {
		fs := afero.NewMemMapFs()
		repo.Fs = fs
		resultChan := make(chan Result, len(tt.hook.Commands)+len(tt.hook.Scripts))
		executor := TestExecutor{}
		runner := &Runner{
			Options: Options{
				Repo:       repo,
				Hook:       tt.hook,
				HookName:   tt.hookName,
				GitArgs:    tt.args,
				ResultChan: resultChan,
			},
			executor: executor,
		}
		gitExec.reset()

		for _, file := range tt.existingFiles {
			if err := fs.MkdirAll(filepath.Dir(file), 0o755); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if err := afero.WriteFile(fs, file, []byte{}, 0o755); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		}

		if len(tt.branch) > 0 {
			if err := afero.WriteFile(fs, filepath.Join(gitPath, "HEAD"), []byte("ref: refs/heads/"+tt.branch), 0o644); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		}

		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			runner.RunAll(context.Background(), tt.sourceDirs)
			close(resultChan)

			var success, fail []Result
			for res := range resultChan {
				if res.Status == StatusOk {
					success = append(success, res)
				} else {
					fail = append(fail, res)
				}
			}

			if !resultsMatch(tt.success, success) {
				t.Errorf("success results are not matching\n Need: %v\n Was: %v", tt.success, success)
			}

			if !resultsMatch(tt.fail, fail) {
				t.Errorf("fail results are not matching:\n Need: %v\n Was: %v", tt.fail, fail)
			}

			if len(gitExec.commands) != len(tt.gitCommands) {
				t.Errorf("wrong git commands\nExpected: %#v\nWas:      %#v", tt.gitCommands, gitExec.commands)
			} else {
				for i, command := range gitExec.commands {
					gitCommandRe := regexp.MustCompile(tt.gitCommands[i])
					if !gitCommandRe.MatchString(command) {
						t.Errorf("wrong git command regexp #%d\nExpected: %s\nWas: %s", i, tt.gitCommands[i], command)
					}
				}
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

func TestReplaceQuoted(t *testing.T) {
	for i, tt := range [...]struct {
		name, source, substitution string
		files                      []string
		result                     string
	}{
		{
			name:         "without substitutions",
			source:       "echo",
			substitution: "{staged_files}",
			files:        []string{"a", "b"},
			result:       "echo",
		},
		{
			name:         "with simple substitution",
			source:       "echo {staged_files}",
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       "echo test.rb README",
		},
		{
			name:         "with single quoted substitution",
			source:       "echo '{staged_files}'",
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       "echo 'test.rb' 'README'",
		},
		{
			name:         "with double quoted substitution",
			source:       `echo "{staged_files}"`,
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       `echo "test.rb" "README"`,
		},
		{
			name:         "with escaped files double quoted",
			source:       `echo "{staged_files}"`,
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo "test me.rb" "README"`,
		},
		{
			name:         "with escaped files single quoted",
			source:       "echo '{staged_files}'",
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo 'test me.rb' 'README'`,
		},
		{
			name:         "with escaped files",
			source:       "echo {staged_files}",
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo 'test me.rb' README`,
		},
		{
			name:         "with many substitutions",
			source:       `echo "{staged_files}" {staged_files}`,
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo "test me.rb" "README" 'test me.rb' README`,
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			result := replaceQuoted(tt.source, tt.substitution, tt.files)
			if result != tt.result {
				t.Errorf("Expected `%s` to eq `%s`", result, tt.result)
			}
		})
	}
}
