package exec

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

const envClicolorForce = "CLICOLOR_FORCE"

// Options contains the data that controls the execution.
type Options struct {
	Root                  string
	Commands              []string
	Env                   map[string]string
	Interactive, UseStdin bool
}

// Executor provides an interface for command execution.
// It is used here for testing purpose mostly.
type Executor interface {
	Execute(context.Context, Options, io.Reader, io.Writer) error
}

func New(logger *logger.ExecutionLogger) Executor {
	return CommandExecutor{logger: logger}
}

// ExitCode returns the exit code of the command that failed with err, if that can
// be determined. It reports false for errors that did not come from a command
// exiting on its own, such as a failure to start the process, or a command
// killed by a signal.
func ExitCode(err error) (int, bool) {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return 0, false
	}

	code := exitErr.ExitCode()
	if code < 0 {
		return 0, false
	}

	return code, true
}

// colorEnv returns an environment variable that propagates lefthook's colors
// setting to the executed command, if there is one to propagate.
func colorEnv(log *logger.ExecutionLogger, opts Options) (string, bool) {
	if log.NoColors() {
		return "NO_COLOR=true", true
	}

	if _, ok := opts.Env[envClicolorForce]; ok {
		return "", false
	}

	if _, ok := os.LookupEnv(envClicolorForce); ok {
		return "", false
	}

	if log.ColorsForced() {
		return envClicolorForce + "=1", true
	}

	return "", false
}
