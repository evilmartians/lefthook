package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	std = New()
)

type Level uint32

const (
	ErrorLevel Level = iota
	InfoLevel
	DebugLevel
)

type Logger struct {
	level Level
	out   io.Writer
	mu    sync.Mutex
}

func New() *Logger {
	return &Logger{
		level: InfoLevel,
		out:   os.Stdout,
	}
}

func Info(args ...interface{}) {
	std.Info(args...)
}

func Debug(args ...interface{}) {
	std.Debug(args...)
}

func Error(args ...interface{}) {
	std.Error(args...)
}

func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

func Println(args ...interface{}) {
	std.Println(args...)
}

func Printf(format string, args ...interface{}) {
	std.Printf(format, args...)
}

func SetLevel(level Level) {
	std.SetLevel(level)
}

func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "error":
		return ErrorLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid Level: %q", lvl)
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(InfoLevel, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(DebugLevel, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(ErrorLevel, format, args...)
}

func (l *Logger) Log(level Level, args ...interface{}) {
	if l.IsLevelEnabled(level) {
		l.Println(args...)
	}
}

func (l *Logger) Logf(level Level, format string, args ...interface{}) {
	if l.IsLevelEnabled(level) {
		l.Printf(format, args...)
	}
}

func (l *Logger) Println(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintln(l.out, args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintf(l.out, format, args...)
}

func (l *Logger) IsLevelEnabled(level Level) bool {
	return l.level >= level
}
