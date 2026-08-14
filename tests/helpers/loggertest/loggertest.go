package loggertest

import (
	"io"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

func New() *logger.Logger {
	// Disable colors by default so test output is deterministic
	// regardless of environment variables such as `NO_COLOR`.
	l := logger.New(io.Discard)
	l.DisableColors()
	return l
}

func NewWithColors() *logger.Logger {
	l := logger.New(io.Discard)
	l.EnableColors()
	return l
}

func NewExecution() *logger.ExecutionLogger {
	return New().NewExecutionLogger()
}
