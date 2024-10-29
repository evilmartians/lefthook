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
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/lefthook/runner/exec"
	"github.com/evilmartians/lefthook/internal/log"
)

type (
	executor struct{}
	cmd      struct{}
	gitCmd   struct {
		mux      sync.Mutex
		commands []string
	}
)

func (e executor) Execute(_ctx context.Context, opts exec.Options, _in io.Reader, _out io.Writer) (err error) {
	if strings.HasPrefix(opts.Commands[0], "success") {
		err = nil
	} else {
		err = errors.New(opts.Commands[0])
	}

	return
}

func (e cmd) RunWithContext(context.Context, []string, string, io.Reader, io.Writer, io.Writer) error {
	return nil
}

func (g *gitCmd) Run(cmd []string, _root string, _in io.Reader, out io.Writer, _errOut io.Writer) error {
	g.mux.Lock()
	g.commands = append(g.commands, strings.Join(cmd, " "))
	g.mux.Unlock()

	cmdLine := strings.Join(cmd, " ")
	if cmdLine == "git diff --name-only --cached --diff-filter=ACMR" ||
		cmdLine == "git diff --name-only HEAD @{push}" {
		root, _ := filepath.Abs("src")
		_, err := out.Write([]byte(strings.Join([]string{
			filepath.Join(root, "scripts", "script.sh"),
			filepath.Join(root, "README.md"),
		}, "\n")))
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *gitCmd) reset() {
	g.mux.Lock()
	g.commands = []string{}
	g.mux.Unlock()
}

func TestRunAll(t *testing.T) {
	root, err := filepath.Abs("src")
	assert.NoError(t, err)

	gitExec := &gitCmd{}
	gitPath := filepath.Join(root, ".git")
	repo := &git.Repository{
		Git:       git.NewExecutor(gitExec),
		HooksPath: filepath.Join(gitPath, "hooks"),
		RootPath:  root,
		GitPath:   gitPath,
		InfoPath:  filepath.Join(gitPath, "info"),
	}

	for name, tt := range map[string]struct {
		branch, hookName string
		args             []string
		sourceDirs       []string
		existingFiles    []string
		hook             *config.Hook
		success, fail    []Result
		gitCommands      []string
		force            bool
		skipLFS          bool
	}{
		"empty hook": {
			hookName: "post-commit",
			hook: &config.Hook{
				Commands: map[string]*config.Command{},
				Scripts:  map[string]*config.Script{},
				Piped:    true,
			},
		},
		"with simple command": {
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
		"with simple command in follow mode": {
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
		"with multiple commands ran in parallel": {
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
		"with exclude tags": {
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
		"with skip=true": {
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
		"with skip=merge": {
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
		"with only=merge match": {
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
		"with only=merge no match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []Result{succeeded("lint")},
		},
		"with hook's skip=merge match": {
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
		"with hook's skip=merge no match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []Result{},
		},
		"with hook's only=merge match": {
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
		"with skip=[merge, rebase] match rebase": {
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
		"with skip=ref match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []Result{},
		},
		"with hook's only=ref match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success: []Result{
				succeeded("lint"),
				succeeded("test"),
			},
		},
		"with hook's only=ref no match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []Result{},
		},
		"with hook's skip=ref no match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success: []Result{
				succeeded("test"),
				succeeded("lint"),
			},
		},
		"with fail": {
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
		"with simple scripts": {
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
		"with simple scripts and only=merge match": {
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
		"with simple scripts and only=merge no match": {
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
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []Result{},
			fail:        []Result{},
		},
		"with interactive=true, parallel=true": {
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
		"with stage_fixed=true": {
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
		"with simple pre-commit": {
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
		"with pre-commit skip": {
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
		"with pre-commit skip but forced": {
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
		"with pre-commit and stage_fixed=true under root": {
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
		"with pre-push skip": {
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
		"with LFS disabled": {
			hookName: "post-checkout",
			skipLFS:  true,
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Commands: map[string]*config.Command{
					"ok": {
						Run: "success",
					},
				},
			},
			success: []Result{succeeded("ok")},
		},
	} {
		fs := afero.NewMemMapFs()
		repo.Fs = fs
		runner := &Runner{
			Options: Options{
				Repo:        repo,
				Hook:        tt.hook,
				HookName:    tt.hookName,
				LogSettings: log.NewSettings(),
				GitArgs:     tt.args,
				Force:       tt.force,
				SkipLFS:     tt.skipLFS,
				SourceDirs:  tt.sourceDirs,
			},
			executor: executor{},
			cmd:      cmd{},
		}
		gitExec.reset()

		for _, file := range tt.existingFiles {
			assert.NoError(t, fs.MkdirAll(filepath.Dir(file), 0o755))
			assert.NoError(t, afero.WriteFile(fs, file, []byte{}, 0o755))
		}

		if len(tt.branch) > 0 {
			assert.NoError(t, afero.WriteFile(fs, filepath.Join(gitPath, "HEAD"), []byte("ref: refs/heads/"+tt.branch), 0o644))
		}

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			git.ResetState()
			results, err := runner.RunAll(context.Background())
			assert.NoError(err)

			var success, fail []Result
			for _, result := range results {
				if result.Success() {
					success = append(success, result)
				} else if result.Failure() {
					fail = append(fail, result)
				}
			}

			assert.ElementsMatch(success, tt.success)
			assert.ElementsMatch(fail, tt.fail)

			assert.Len(gitExec.commands, len(tt.gitCommands))
			for i, command := range gitExec.commands {
				gitCommandRe := regexp.MustCompile(tt.gitCommands[i])
				if !gitCommandRe.MatchString(command) {
					t.Errorf("wrong git command regexp #%d\nExpected: %s\nWas: %s", i, tt.gitCommands[i], command)
				}
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
			assert.Equal(t, tt.result, tt.names)
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
			assert.Equal(t, tt.result, tt.names)
		})
	}
}
