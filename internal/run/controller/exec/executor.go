package exec

import (
	"context"
	"io"
	"os"
	"runtime"
	"strings"

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

	if hasEnvKey(opts.Env, envClicolorForce, runtime.GOOS == "windows") {
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

// hasEnvKey reports whether env sets key. Windows treats variable names
// case-insensitively, so a job's clicolor_force is the same variable there.
func hasEnvKey(env map[string]string, key string, caseInsensitive bool) bool {
	if _, ok := env[key]; ok {
		return true
	}
	if !caseInsensitive {
		return false
	}
	for name := range env {
		if strings.EqualFold(name, key) {
			return true
		}
	}
	return false
}
