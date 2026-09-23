package runner

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/runner/executor"
	"github.com/evilmartians/lefthook/v2/internal/runner/result"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/configtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

// testHookName is a hook that does not use staged files, so jobs are not skipped.
const testHookName = "commit-msg"

type funcExecutor func(ctx context.Context, opts executor.Options, in io.Reader, out io.Writer) error

func (f funcExecutor) Execute(ctx context.Context, opts executor.Options, in io.Reader, out io.Writer) error {
	return f(ctx, opts, in, out)
}

func newTestRunner(exec executor.Executor, stdin io.Reader) *Runner {
	repo := gittest.NewRepositoryBuilder().
		Root("/src").
		Cmd(cmdtest.NewTracking(nil)).
		Fs(afero.NewMemMapFs()).
		Build()
	repo.ResetCache()

	log := loggertest.NewExecution()

	return &Runner{
		git:          repo,
		logger:       log,
		stdin:        newStdin(stdin, log),
		executor:     exec,
		cmd:          cmdtest.NewTracking(nil),
		filesToStage: newStageFilesList(),
	}
}

func TestRunHook_ParallelStdin(t *testing.T) {
	const input = "refs/heads/main 1111 refs/heads/main 2222\n"

	for name, tt := range map[string]struct {
		hook string
		want map[string]string
	}{
		"parallel jobs": {
			hook: `
        parallel: true
        jobs:
          - name: a
            run: a
            use_stdin: true
          - name: b
            run: b
            use_stdin: true
          - name: c
            run: c
            use_stdin: true
      `,
			want: map[string]string{"a": input, "b": input, "c": input},
		},
		"sequential job reads part of the input": {
			hook: `
        jobs:
          - name: head
            run: head
            use_stdin: true
          - name: all
            run: all
            use_stdin: true
      `,
			want: map[string]string{"head": input[:4], "all": input},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			got := make(map[string]string)

			exec := funcExecutor(func(_ context.Context, opts executor.Options, in io.Reader, _ io.Writer) error {
				cmd := strings.TrimSpace(opts.Commands[0])

				var data []byte
				if cmd == "head" {
					data = make([]byte, 4)
					n, err := in.Read(data)
					if err != nil {
						return err
					}
					data = data[:n]
				} else {
					var err error
					data, err = io.ReadAll(in)
					if err != nil {
						return err
					}
				}

				mu.Lock()
				defer mu.Unlock()
				got[cmd] = string(data)

				return nil
			})

			hook := configtest.ParseHook(tt.hook)
			hook.Name = testHookName

			runner := newTestRunner(exec, strings.NewReader(input))
			_, err := runner.RunHook(t.Context(), Options{DisableTTY: true}, hook)
			assert.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunHook_ParallelResultsOrder(t *testing.T) {
	for name, tt := range map[string]struct {
		hook string
		want []string
	}{
		"top-level jobs": {
			hook: `
        parallel: true
        jobs:
          - name: slow
            run: "60"
          - name: medium
            run: "30"
          - name: fast
            run: "0"
      `,
			want: []string{"slow", "medium", "fast"},
		},
		"group jobs": {
			hook: `
        jobs:
          - name: group
            group:
              parallel: true
              jobs:
                - name: slow
                  run: "60"
                - name: fast
                  run: "0"
      `,
			want: []string{"slow", "fast"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// The command is a delay in milliseconds: jobs finish in reverse order.
			exec := funcExecutor(func(ctx context.Context, opts executor.Options, _ io.Reader, _ io.Writer) error {
				delay, err := time.ParseDuration(strings.TrimSpace(opts.Commands[0]) + "ms")
				if err != nil {
					return err
				}

				select {
				case <-time.After(delay):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			hook := configtest.ParseHook(tt.hook)
			hook.Name = testHookName

			runner := newTestRunner(exec, strings.NewReader(""))
			results, err := runner.RunHook(t.Context(), Options{DisableTTY: true}, hook)
			assert.NoError(t, err)

			if len(results) == 1 && len(results[0].Sub) > 0 {
				results = results[0].Sub
			}

			names := make([]string, 0, len(results))
			for _, res := range results {
				names = append(names, res.Name)
			}
			assert.Equal(t, tt.want, names)
		})
	}
}

func TestRunHook_Timeout(t *testing.T) {
	for name, tt := range map[string]struct {
		hook        string
		ctxTimeout  time.Duration
		wantFailure result.Result
	}{
		"job timeout": {
			hook: `
        jobs:
          - name: slow
            run: slow
            timeout: 20ms
      `,
			wantFailure: failed("slow", "timeout (20ms)"),
		},
		"parent deadline is not a job timeout": {
			hook: `
        jobs:
          - name: slow
            run: slow
            timeout: 10s
            fail_text: interrupted
      `,
			ctxTimeout:  20 * time.Millisecond,
			wantFailure: failed("slow", "interrupted"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			exec := funcExecutor(func(ctx context.Context, _ executor.Options, _ io.Reader, _ io.Writer) error {
				<-ctx.Done()
				return ctx.Err()
			})

			ctx := t.Context()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			hook := configtest.ParseHook(tt.hook)
			hook.Name = testHookName

			runner := newTestRunner(exec, strings.NewReader(""))
			results, err := runner.RunHook(ctx, Options{DisableTTY: true}, hook)
			assert.NoError(t, err)

			if assert.Len(t, results, 1) {
				assert.Equal(t, tt.wantFailure, failed(results[0].Name, results[0].Text()))
			}
		})
	}
}
