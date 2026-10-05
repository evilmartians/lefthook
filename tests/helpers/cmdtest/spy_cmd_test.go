package cmdtest_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
)

var (
	_ system.Command            = (*cmdtest.SpyCmd)(nil)
	_ system.CommandWithContext = (*cmdtest.SpyCmd)(nil)
)

func TestSpyCmd_Run(t *testing.T) {
	errFailed := errors.New("failed")

	for name, tt := range map[string]struct {
		callback   func(string, string, io.Writer) error
		wantOutput string
		wantErr    error
	}{
		"nil-callback": {},
		"callback-output": {
			callback: func(cmd string, root string, out io.Writer) error {
				_, err := out.Write([]byte(root + ": " + cmd))
				return err
			},
			wantOutput: "root: A 1",
		},
		"callback-error": {
			callback: func(string, string, io.Writer) error {
				return errFailed
			},
			wantErr: errFailed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewSpyCmd(tt.callback)
			out := new(bytes.Buffer)

			err := cmd.Run([]string{"A", "1"}, "root", system.NullReader, out, io.Discard)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SpyCmd.Run() error = %v, want %v", err, tt.wantErr)
			}

			if out.String() != tt.wantOutput {
				t.Errorf("SpyCmd.Run() output = %q, want %q", out.String(), tt.wantOutput)
			}

			if want := []string{"A 1"}; !cmp.Equal(cmd.Commands, want) {
				t.Errorf("SpyCmd.Commands = %v, want %v", cmd.Commands, want)
			}
		})
	}
}

func TestSpyCmd_RunWithContext(t *testing.T) {
	var called []string
	cmd := cmdtest.NewSpyCmd(func(cmd string, _ string, _ io.Writer) error {
		called = append(called, cmd)
		return nil
	})

	_ = cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard)
	err := cmd.RunWithContext(t.Context(), []string{"B", "2"}, "", system.NullReader, io.Discard, io.Discard)
	if err != nil {
		t.Errorf("SpyCmd.RunWithContext() error = %v, want nil", err)
	}

	want := []string{"A 1", "B 2"}
	if !cmp.Equal(cmd.Commands, want) {
		t.Errorf("SpyCmd.Commands = %v, want %v", cmd.Commands, want)
	}

	if !cmp.Equal(called, want) {
		t.Errorf("SpyCmd callback calls = %v, want %v", called, want)
	}
}

func TestSpyCmd_Reset(t *testing.T) {
	cmd := cmdtest.NewSpyCmd(nil)
	_ = cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard)

	cmd.Reset()

	if len(cmd.Commands) != 0 {
		t.Errorf("SpyCmd.Commands = %v, want empty", cmd.Commands)
	}
}

func TestSpyCmd_WithoutEnvs(t *testing.T) {
	cmd := cmdtest.NewSpyCmd(nil)

	if result := cmd.WithoutEnvs("GIT_DIR"); result != cmd {
		t.Errorf("SpyCmd.WithoutEnvs() = %v, want %v", result, cmd)
	}
}
