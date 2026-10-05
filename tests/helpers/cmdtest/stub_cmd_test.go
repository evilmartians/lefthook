package cmdtest_test

import (
	"bytes"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
)

var _ system.Command = (*cmdtest.StubCmd)(nil)

func TestStubCmd_Run(t *testing.T) {
	cmd := cmdtest.NewStubCmd()
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)

	err := cmd.Run([]string{"git", "status"}, "root", system.NullReader, out, errOut)
	if err != nil {
		t.Errorf("StubCmd.Run() error = %v, want nil", err)
	}

	if out.Len() != 0 || errOut.Len() != 0 {
		t.Errorf("StubCmd.Run() output = %q, %q, want empty", out.String(), errOut.String())
	}
}

func TestStubCmd_WithoutEnvs(t *testing.T) {
	cmd := cmdtest.NewStubCmd()

	if result := cmd.WithoutEnvs("GIT_DIR"); result != cmd {
		t.Errorf("StubCmd.WithoutEnvs() = %v, want %v", result, cmd)
	}
}
