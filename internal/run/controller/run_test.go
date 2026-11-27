package controller

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/exec"
)

type outputExecutor struct {
	output string
}

func (e outputExecutor) Execute(_ context.Context, _ exec.Options, _ io.Reader, out io.Writer) error {
	_, err := io.WriteString(out, e.output)
	return err
}

func TestController_run(t *testing.T) {
	for name, tt := range map[string]struct {
		output string
		want   string
	}{
		"no-output": {
			output: "",
			want:   "",
		},
		"with-output": {
			output: "done",
			want:   "done\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			log := logger.New(&buf)
			log.DisableColors()

			controller := &Controller{
				logger:   log.NewExecutionLogger("execution_out"),
				executor: outputExecutor{output: tt.output},
			}

			err := controller.run(t.Context(), "job", false, exec.Options{Commands: []string{"job"}})
			if err != nil {
				t.Fatalf("Controller.run() error = %v", err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("Controller.run() printed %q, want %q", got, tt.want)
			}
		})
	}
}
