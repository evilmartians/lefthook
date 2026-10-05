package cmdtest_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
)

var _ system.Command = (*cmdtest.FakeCmd)(nil)

// recordTB records errors instead of failing the test.
type recordTB struct {
	testing.TB

	errors []string
}

func (r *recordTB) Helper() {}

func (r *recordTB) Errorf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func TestFakeCmd_Run(t *testing.T) {
	errFailed := errors.New("failed")

	for name, tt := range map[string]struct {
		outs       []cmdtest.Out
		command    []string
		wantOutput string
		wantErr    error
		wantErrors int
	}{
		"matching-command": {
			outs:       []cmdtest.Out{{Command: "git status", Output: "clean"}},
			command:    []string{"git", "status"},
			wantOutput: "clean",
		},
		"returns-error": {
			outs:    []cmdtest.Out{{Command: "git status", Err: errFailed}},
			command: []string{"git", "status"},
			wantErr: errFailed,
		},
		"wrong-command": {
			outs:       []cmdtest.Out{{Command: "git status", Output: "clean"}},
			command:    []string{"git", "diff"},
			wantOutput: "clean",
			wantErrors: 1,
		},
		"no-commands-left": {
			command:    []string{"git", "status"},
			wantErrors: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			tb := &recordTB{TB: t}
			cmd := cmdtest.NewFakeCmd(tb, tt.outs)
			out := new(bytes.Buffer)

			err := cmd.Run(tt.command, "", system.NullReader, out, io.Discard)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FakeCmd.Run() error = %v, want %v", err, tt.wantErr)
			}

			if out.String() != tt.wantOutput {
				t.Errorf("FakeCmd.Run() output = %q, want %q", out.String(), tt.wantOutput)
			}

			if len(tb.errors) != tt.wantErrors {
				t.Errorf("FakeCmd.Run() reported %v, want %d errors", tb.errors, tt.wantErrors)
			}
		})
	}
}

func TestFakeCmd_Run_order(t *testing.T) {
	tb := &recordTB{TB: t}
	cmd := cmdtest.NewFakeCmd(tb, []cmdtest.Out{
		{Command: "A 1"},
		{Command: "B 2"},
	})

	_ = cmd.Run([]string{"B", "2"}, "", system.NullReader, io.Discard, io.Discard)
	_ = cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard)

	if len(tb.errors) != 2 {
		t.Errorf("FakeCmd.Run() reported %v, want 2 errors", tb.errors)
	}
}

func TestFakeCmd_WithoutEnvs(t *testing.T) {
	cmd := cmdtest.NewFakeCmd(t, nil)

	if result := cmd.WithoutEnvs("GIT_DIR"); result != cmd {
		t.Errorf("FakeCmd.WithoutEnvs() = %v, want %v", result, cmd)
	}
}
