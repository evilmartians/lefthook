package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
)

const (
	envVerbose       = "LEFTHOOK_VERBOSE"
	envForceColor    = "FORCE_COLOR"
	envNoColor       = "NO_COLOR"
	envClicolorForce = "CLICOLOR_FORCE"
	envClicolor      = "CLICOLOR"
)

type Level uint8

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var border = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderLeft(true).
	PaddingLeft(2) //nolint:mnd

type Logger struct {
	mu           sync.Mutex
	level        Level
	out          io.Writer
	colors       ColorsSetting
	colorsForced bool

	Spinner *Spinner
}

func New(out io.Writer) *Logger {
	return &Logger{
		out:     out,
		level:   defaultLogLevel(),
		colors:  defaultColors(),
		Spinner: NewSpinner(),
	}
}

func (l *Logger) Error(args ...any) {
	l.log(LevelError, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...any) {
	l.log(LevelWarn, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...any) {
	l.log(LevelInfo, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(args ...any) {
	l.log(LevelDebug, args...)
}

func (l *Logger) log(level Level, args ...any) {
	if l.level < level {
		return
	}

	var message string
	switch level {
	case LevelDebug:
		strArgs := make([]string, 0, len(args))
		for _, arg := range args {
			strArgs = append(strArgs, l.Paint(ColorGray, fmt.Sprintf("%v", arg)))
		}
		message = border.BorderForeground(l.colors.get(ColorGray)).Render(strArgs...)
	case LevelWarn:
		strArgs := make([]string, 0, len(args))
		for _, arg := range args {
			strArgs = append(strArgs, l.Paint(ColorYellow, fmt.Sprintf("%v", arg)))
		}
		message = border.BorderForeground(l.colors.get(ColorYellow)).Render(strArgs...)
	case LevelError:
		strArgs := make([]string, 0, len(args))
		for _, arg := range args {
			strArgs = append(strArgs, l.Paint(ColorRed, fmt.Sprintf("%v", arg)))
		}

		message = border.BorderForeground(l.colors.get(ColorRed)).Render(strArgs...)
	case LevelInfo:
		message = fmt.Sprint(args...)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Spinner.active() {
		l.Spinner.Stop()
		defer l.Spinner.Start()
	}

	_, _ = fmt.Fprintln(l.out, message)
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) NoColors() bool {
	return l.colors.kind == colorsDisabled
}

func isEnvEnabled(env string) bool {
	value := os.Getenv(env)
	if len(value) > 0 && value != "0" && value != "false" && value != "off" {
		return true
	}

	return false
}

func isEnvDisabled(env string) bool {
	value := os.Getenv(env)
	if value == "0" || value == "false" || value == "off" {
		return true
	}

	return false
}

func defaultLogLevel() Level {
	if isEnvEnabled(envVerbose) {
		return LevelDebug
	}

	return LevelInfo
}

func defaultColors() ColorsSetting {
	if isEnvEnabled(envForceColor) {
		return DefaultColors
	}

	if isEnvEnabled(envClicolorForce) {
		return DefaultColors
	}

	if isEnvEnabled(envNoColor) {
		return NoColors
	}

	if isEnvDisabled(envClicolor) {
		return NoColors
	}

	return DefaultColors
}
