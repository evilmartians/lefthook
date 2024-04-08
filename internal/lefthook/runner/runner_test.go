package runner

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
	"github.com/evilmartians/lefthook/internal/lefthook/runner/exec"
	"github.com/evilmartians/lefthook/internal/log"
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

func (g *GitMock) Execute(cmd []string, _root string) (string, error) {
	g.mux.Lock()
	g.commands = append(g.commands, strings.Join(cmd, " "))
	g.mux.Unlock()

	cmdLine := strings.Join(cmd, " ")
	if cmdLine == "git diff --name-only --cached --diff-filter=ACMR" ||
		cmdLine == "git diff --name-only HEAD @{push}" {
		root, _ := filepath.Abs("src")
		return strings.Join([]string{
			filepath.Join(root, "scripts", "script.sh"),
			filepath.Join(root, "README.md"),
		}, "\n"), nil
	}

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
		Git:       git.NewExecutor(gitExec),
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
		force                  bool
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
			success: []Result{succeeded("test")},
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
			success: []Result{succeeded("test")},
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
				succeeded("test"),
				succeeded("lint"),
			},
			fail: []Result{failed("type-check", "")},
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
			success: []Result{succeeded("lint")},
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
			success: []Result{succeeded("lint")},
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
			success: []Result{succeeded("lint")},
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
				succeeded("lint"),
				succeeded("test"),
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
			success: []Result{succeeded("lint")},
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
				succeeded("lint"),
				succeeded("test"),
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
			success: []Result{succeeded("lint")},
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
				succeeded("lint"),
				succeeded("test"),
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
				succeeded("test"),
				succeeded("lint"),
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
			fail: []Result{failed("test", "try 'success'")},
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
			success: []Result{succeeded("script.sh")},
			fail:    []Result{failed("failing.js", "install node")},
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
			success: []Result{succeeded("script.sh")},
			fail:    []Result{failed("failing.js", "install node")},
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
			success: []Result{}, // script.sh and ok are skipped because of non-interactive cmd failure
			fail:    []Result{failed("failing.js", ""), failed("fail", "")},
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
			success: []Result{succeeded("ok"), succeeded("success.sh")},
			fail:    []Result{failed("fail", ""), failed("failing.js", "")},
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
			success: []Result{succeeded("ok"), succeeded("success.sh")},
			fail:    []Result{failed("fail", ""), failed("failing.js", "")},
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
			success: []Result{succeeded("ok")},
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
			name:     "skippable pre-commit hook with force",
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
			force:   true,
			success: []Result{succeeded("ok")},
			fail:    []Result{failed("fail", "")},
			gitCommands: []string{
				"git status --short",
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
			success: []Result{succeeded("ok")},
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
			success: []Result{succeeded("ok")},
			gitCommands: []string{
				"git diff --name-only HEAD @{push}",
				"git diff --name-only HEAD @{push}",
			},
		},
	} {
		fs := afero.NewMemMapFs()
		repo.Fs = fs
		executor := TestExecutor{}
		runner := &Runner{
			Options: Options{
				Repo:        repo,
				Hook:        tt.hook,
				HookName:    tt.hookName,
				LogSettings: log.NewSettings(),
				GitArgs:     tt.args,
				Force:       tt.force,
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
			results := runner.RunAll(context.Background(), tt.sourceDirs)

			var success, fail []Result
			for _, result := range results {
				if result.Success() {
					success = append(success, result)
				} else if result.Failure() {
					fail = append(fail, result)
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
		str := fmt.Sprintf("%v", item)
		matches[str] = struct{}{}
	}

	for _, item := range b {
		str := fmt.Sprintf("%v", item)
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

//nolint:dupl
func TestSortByPriorityCommands(t *testing.T) {
	for i, tt := range [...]struct {
		name     string
		names    []string
		commands map[string]*config.Command
		result   []string
	}{
		{
			name:     "alphanumeric sort",
			names:    []string{"10_a", "1_a", "2_a", "5_a"},
			commands: map[string]*config.Command{},
			result:   []string{"1_a", "2_a", "5_a", "10_a"},
		},
		{
			name:  "partial priority",
			names: []string{"10_a", "1_a", "2_a", "5_a"},
			commands: map[string]*config.Command{
				"5_a":  {Priority: 10},
				"2_a":  {Priority: 1},
				"10_a": {},
			},
			result: []string{"2_a", "5_a", "1_a", "10_a"},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i+1, tt.name), func(t *testing.T) {
			sortByPriority(tt.names, tt.commands)
			for i, name := range tt.result {
				if tt.names[i] != name {
					t.Errorf("Not matching on index %d: %s != %s", i, name, tt.names[i])
				}
			}
		})
	}
}

//nolint:dupl
func TestSortByPriorityScripts(t *testing.T) {
	for i, tt := range [...]struct {
		name    string
		names   []string
		scripts map[string]*config.Script
		result  []string
	}{
		{
			name:    "alphanumeric sort",
			names:   []string{"10_a.sh", "1_a.sh", "2_a.sh", "5_b.sh"},
			scripts: map[string]*config.Script{},
			result:  []string{"1_a.sh", "2_a.sh", "5_b.sh", "10_a.sh"},
		},
		{
			name:  "partial priority",
			names: []string{"10.rb", "file.sh", "script.go", "5_a.sh"},
			scripts: map[string]*config.Script{
				"5_a.sh":    {Priority: 10},
				"script.go": {Priority: 1},
				"10.rb":     {},
			},
			result: []string{"script.go", "5_a.sh", "10.rb", "file.sh"},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i+1, tt.name), func(t *testing.T) {
			sortByPriority(tt.names, tt.scripts)
			for i, name := range tt.result {
				if tt.names[i] != name {
					t.Errorf("Not matching on index %d: %s != %s", i, name, tt.names[i])
				}
			}
		})
	}
}
