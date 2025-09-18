package controller

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/internal/run/result"
	"github.com/evilmartians/lefthook/internal/system"
	"github.com/evilmartians/lefthook/tests/helpers"
	"github.com/evilmartians/lefthook/tests/helpers/git"
)

type (
	executor struct{}
	cmd      struct{}
	gitCmd   struct {
		mux      sync.Mutex
		commands []string
	}
)

func succeeded(name string) result.Result {
	return result.Success(name, time.Second)
}

func failed(name, failText string) result.Result {
	return result.Failure(name, failText, time.Second)
}

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

func (g *gitCmd) WithoutEnvs(...string) system.Command {
	return g
}

func (g *gitCmd) Run(cmd []string, _root string, _in io.Reader, out io.Writer, _errOut io.Writer) error {
	g.mux.Lock()
	g.commands = append(g.commands, strings.Join(cmd, " "))
	g.mux.Unlock()

	cmdLine := strings.Join(cmd, " ")
	if cmdLine == "git diff --name-only --cached --diff-filter=ACMR" ||
		cmdLine == "git diff --name-only --cached --diff-filter=ACMRD" ||
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
	gitPath := git.GitPath(root)

	for name, tt := range map[string]struct {
		branch, hookName string
		args             []string
		sourceDirs       []string
		existingFiles    []string
		hook             *config.Hook
		success, fail    []result.Result
		gitCommands      []string
		force            bool
		skipLFS          bool
	}{
		"empty hook": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        piped: true
      `),
		},
		"with simple command": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
      `),
			success: []result.Result{succeeded("test")},
		},
		"with simple command in follow mode": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        follow: true
        jobs:
          - name: test
            run: "success"
      `),
			success: []result.Result{succeeded("test")},
		},
		"with multiple commands ran in parallel": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
          - name: type-check
            run: fail
      `),
			success: []result.Result{
				succeeded("test"),
				succeeded("lint"),
			},
			fail: []result.Result{failed("type-check", "")},
		},
		"with exclude tags": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        exclude_tags: [test, formatter]
        jobs:
          - name: test
            run: success
          - name: formatter
            run: success
          - name: lint
            run: success
      `),
			success: []result.Result{succeeded("lint")},
		},
		"with skip=true": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
            skip: true
          - name: lint
            run: success
      `),
			success: []result.Result{succeeded("lint")},
		},
		"with skip=merge": {
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
            skip: merge
          - name: lint
            run: success
      `),
			success: []result.Result{succeeded("lint")},
		},
		"with only=merge match": {
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
            only: merge
          - name: lint
            run: success
            skip: merge
      `),
			success: []result.Result{
				succeeded("test"),
			},
		},
		"with only=merge no match": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
            only: merge
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []result.Result{succeeded("lint")},
		},
		"with hook's skip=merge match": {
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: helpers.ParseHook(`
        skip: merge
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			success: []result.Result{},
		},
		"with hook's only=merge no match": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        only: merge
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []result.Result{},
		},
		"with hook's only=merge match": {
			hookName: "post-commit",
			existingFiles: []string{
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hook: helpers.ParseHook(`
        only: merge
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			success: []result.Result{
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
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: success
            skip:
              - merge
              - rebase
          - name: lint
            run: success
      `),
			success: []result.Result{succeeded("lint")},
		},
		"with skip=ref match": {
			branch: "main",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        skip:
          - merge
          - ref: main
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []result.Result{},
		},
		"with hook's only=ref match": {
			branch: "main",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        only:
          - merge
          - ref: main
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success: []result.Result{
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
			hook: helpers.ParseHook(`
        only:
          - merge
          - ref: main
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []result.Result{},
		},
		"with hook's skip=ref no match": {
			branch: "fix",
			existingFiles: []string{
				filepath.Join(gitPath, "HEAD"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        skip:
          - merge
          - ref: main
        jobs:
          - name: test
            run: success
          - name: lint
            run: success
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success: []result.Result{
				succeeded("test"),
				succeeded("lint"),
			},
		},
		"with fail": {
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: test
            run: fail
            fail_text: try 'success'
      `),
			fail: []result.Result{failed("test", "try 'success'")},
		},
		"with simple scripts": {
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - script: "script.sh"
            runner: success
          - script: "failing.js"
            runner: fail
            fail_text: install node
      `),
			success: []result.Result{succeeded("script.sh")},
			fail:    []result.Result{failed("failing.js", "install node")},
		},
		"with simple scripts and only=merge match": {
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
				filepath.Join(gitPath, "MERGE_HEAD"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - script: "script.sh"
            runner: success
            only: merge
          - script: "failing.js"
            only: merge
            runner: fail
            fail_text: install node
      `),
			success: []result.Result{succeeded("script.sh")},
			fail:    []result.Result{failed("failing.js", "install node")},
		},
		"with simple scripts and only=merge no match": {
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - script: "script.sh"
            runner: success
            only: merge
          - script: "failing.js"
            only: merge
            runner: fail
            fail_text: install node
      `),
			gitCommands: []string{`git show --no-patch --format="%P"`},
			success:     []result.Result{},
			fail:        []result.Result{},
		},
		"with interactive=true, parallel=true": {
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "script.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        parallel: true
        jobs:
          - name: ok
            run: success
            interactive: true
          - name: fail
            run: fail
          - script: "script.sh"
            runner: success
            interactive: true
          - script: "failing.js"
            runner: fail
      `),
			success: []result.Result{succeeded("ok"), succeeded("script.sh")},
			fail:    []result.Result{failed("failing.js", ""), failed("fail", "")},
		},
		"with stage_fixed=true": {
			sourceDirs: []string{filepath.Join(root, config.DefaultSourceDir)},
			existingFiles: []string{
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "success.sh"),
				filepath.Join(root, config.DefaultSourceDir, "post-commit", "failing.js"),
			},
			hookName: "post-commit",
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
            stage_fixed: true
          - name: fail
            run: fail
            stage_fixed: true
          - script: "success.sh"
            runner: success
            stage_fixed: true
          - script: "failing.js"
            runner: fail
            stage_fixed: true
      `),
			success: []result.Result{succeeded("ok"), succeeded("success.sh")},
			fail:    []result.Result{failed("fail", ""), failed("failing.js", "")},
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
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
            stage_fixed: true
          - name: fail
            run: fail
            stage_fixed: true
          - script: "success.sh"
            runner: success
            stage_fixed: true
          - script: "failing.js"
            runner: fail
            stage_fixed: true
      `),
			success: []result.Result{succeeded("ok"), succeeded("success.sh")},
			fail:    []result.Result{failed("fail", ""), failed("failing.js", "")},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add --force .*script.sh.*README.md",
				"git add --force .*script.sh.*README.md",
			},
		},
		"with pre-commit skip": {
			hookName: "pre-commit",
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
            stage_fixed: true
            glob:
              - "*.md"
          - name: fail
            run: fail
            stage_fixed: true
            glob:
              - "*.txt"
      `),
			success: []result.Result{succeeded("ok")},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMRD",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add --force .*README.md",
			},
		},
		"with pre-commit skip but forced": {
			hookName: "pre-commit",
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
            stage_fixed: true
            glob:
              - "*.md"
          - name: fail
            run: fail
            stage_fixed: true
            glob:
              - "*.sh"
      `),
			force:   true,
			success: []result.Result{succeeded("ok")},
			fail:    []result.Result{failed("fail", "")},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add --force .*README.md",
			},
		},
		"with pre-commit and stage_fixed=true under root": {
			hookName: "pre-commit",
			existingFiles: []string{
				filepath.Join(root, "scripts", "script.sh"),
				filepath.Join(root, "README.md"),
			},
			hook: &config.Hook{
				Jobs: []*config.Job{{
					Name:       "ok",
					Run:        "success",
					Root:       filepath.Join(root, "scripts"),
					StageFixed: true,
				}},
			},
			success: []result.Result{succeeded("ok")},
			gitCommands: []string{
				"git status --short",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git diff --name-only --cached --diff-filter=ACMR",
				"git add --force .*scripts.*script.sh",
			},
		},
		"with pre-push skip": {
			hookName: "pre-push",
			existingFiles: []string{
				filepath.Join(root, "README.md"),
			},
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
            stage_fixed: true
            glob:
              - "*.md"
          - name: fail
            run: fail
            stage_fixed: true
            glob:
              - "*.sh"
      `),
			success: []result.Result{succeeded("ok")},
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
			hook: helpers.ParseHook(`
        jobs:
          - name: ok
            run: success
      `),
			success: []result.Result{succeeded("ok")},
		},
	} {
		fs := afero.NewMemMapFs()
		repo := git.NewRepositoryBuilder().Root(root).Git(gitExec).Fs(fs).Build()
		controller := &Controller{
			git:      repo,
			executor: executor{},
			cmd:      cmd{},
		}
		gitExec.reset()

		for _, file := range tt.existingFiles {
			assert.NoError(t, fs.MkdirAll(filepath.Dir(file), 0o755))
			assert.NoError(t, afero.WriteFile(fs, file, []byte{}, 0o755))
		}

		if len(tt.branch) > 0 {
			assert.NoError(t, afero.WriteFile(fs, filepath.Join(repo.GitPath, "HEAD"), []byte("ref: refs/heads/"+tt.branch), 0o644))
		}

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			repo.Setup()
			gitExec.reset()

			opts := Options{
				GitArgs:    tt.args,
				Force:      tt.force,
				SkipLFS:    tt.skipLFS,
				SourceDirs: tt.sourceDirs,
			}
			tt.hook.Name = tt.hookName
			results, err := controller.RunHook(t.Context(), opts, tt.hook)
			assert.NoError(err)

			var success, fail []result.Result
			for _, result := range results {
				if result.Success() {
					success = append(success, succeeded(result.Name))
				} else if result.Failure() {
					fail = append(fail, failed(result.Name, result.Text()))
				}
			}

			assert.ElementsMatch(success, tt.success)
			assert.ElementsMatch(fail, tt.fail)

			if len(tt.gitCommands) > 0 {
				assert.Len(gitExec.commands, len(tt.gitCommands))
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
