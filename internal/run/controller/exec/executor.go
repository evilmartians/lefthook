package exec

import (
	"context"
	"io"
	"os"

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
