package loggertest

import (
	"io"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

func New() *logger.Logger {
	return logger.New(io.Discard)
}

func NewExecution() *logger.ExecutionLogger {
	return logger.New(io.Discard).NewExecutionLogger()
}
