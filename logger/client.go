package logger

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// Logger instance with log client
type Logger struct {
	*logrus.Logger
}

var (
	once   sync.Once
	logger *Logger
)

const (
	Info  = logrus.InfoLevel
	Warn  = logrus.WarnLevel
	Debug = logrus.DebugLevel
	Trace = logrus.TraceLevel
	Error = logrus.ErrorLevel
	Fatal = logrus.FatalLevel
	Panic = logrus.PanicLevel
)

// SetOutput sets logger output (uses for capturing output in tests)
func (l *Logger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

// Reset removes existing instance of sync.once for re-init the current Logger instance after GetInstance() call
func (l *Logger) Reset() {
	once = *new(sync.Once)
}

func TransformLogLevel(logLevel string) (logrus.Level, error) {
	return logrus.ParseLevel(logLevel)
}

// GetInstance return singleton logger instance
func GetInstance(logLevel logrus.Level) *Logger {
	once.Do(func() {
		logrusLog := logrus.New()
		logrusLog.SetLevel(logLevel)
		logrusLog.SetFormatter(&logrus.TextFormatter{})
		// create instance of logger
		logger = &Logger{logrusLog}
		// set logger output
		logger.SetOutput(os.Stdout)
	})

	return logger
}

// Log writes simple log
func (l *Logger) Log(level logrus.Level, message string) {
	l.Logger.Log(level, message)
}
