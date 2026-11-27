//go:build windows

package exec_test

import (
	"bytes"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestCommandExecutor_Execute(t *testing.T) {
	for name, tt := range map[string]struct {
		useStdin bool
		want     string
	}{
		"stderr-in-output": {
			want: "stdout-line\nstderr-line\n",
		},
		"stdin-keeps-console": {
			useStdin: true,
			want:     "stdout-line\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			out := new(bytes.Buffer)
			opts := exec.Options{
				Root:     t.TempDir(),
				Commands: []string{"echo stdout-line; echo stderr-line >&2"},
				UseStdin: tt.useStdin,
			}

			err := exec.New(loggertest.NewExecution()).Execute(t.Context(), opts, system.NullReader, out)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if got := out.String(); got != tt.want {
				t.Errorf("Execute() output = %q, want %q", got, tt.want)
			}
		})
	}
}
