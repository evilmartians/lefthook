package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/logrusorgru/aurora"
)

const grayColor = 11

var std = New()

type Level uint32

const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

type Logger struct {
	level  Level
	aurora aurora.Aurora
	out    io.Writer
	mu     sync.Mutex
}

func New() *Logger {
	return &Logger{
		level:  InfoLevel,
		out:    os.Stdout,
		aurora: aurora.NewAurora(true),
	}
}

func Debug(args ...interface{}) {
	res := fmt.Sprint(args...)
	std.Debug(std.aurora.Gray(grayColor, res))
}

func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	std.Info(args...)
}

func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

func Error(args ...interface{}) {
	res := fmt.Sprint(args...)
	std.Error(Red(res))
}

func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

func Warn(args ...interface{}) {
	res := fmt.Sprint(args...)
	std.Warn(std.aurora.Yellow(res))
}

func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
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

func SetColors(enable bool) {
	std.aurora = aurora.NewAurora(enable)
}

func Cyan(arg interface{}) aurora.Value {
	return std.aurora.Cyan(arg)
}

func Green(arg interface{}) aurora.Value {
	return std.aurora.Green(arg)
}

func Red(arg interface{}) aurora.Value {
	return std.aurora.Red(arg)
}

func Bold(arg interface{}) aurora.Value {
	return std.aurora.Bold(arg)
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

func (l *Logger) Warn(args ...interface{}) {
	l.Log(WarnLevel, args...)
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

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logf(WarnLevel, format, args...)
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
