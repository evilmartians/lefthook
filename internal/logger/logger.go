package logger

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
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

var profile = colorprofile.Detect(os.Stdout, os.Environ())
var complete = lipgloss.Complete(profile)
var border = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderLeft(true).
	PaddingLeft(2)

type Logger struct {
	mu     sync.Mutex
	level  Level
	out    io.Writer
	colors map[Color]color.Color

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

func (l *Logger) Error(args ...string) {
	message := border.BorderForeground(l.colors[ColorRed]).Render(args...)
	l.log(LevelError, message)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...string) {
	message := border.BorderForeground(l.colors[ColorYellow]).Render(args...)
	l.log(LevelWarn, message)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...any) {
	l.log(LevelInfo, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(args ...string) {
	message := border.BorderForeground(l.colors[ColorGray]).Render(args...)
	l.log(LevelDebug, message)
}

func (l *logger) log(level Level, args ...any) {
	if l.level < level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Spinner.active() {
		l.Spinner.Stop()
		defer l.Spinner.Start()
	}

	_, _ = fmt.Fprintln(l.out, args...)
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
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

func defaultColors() map[Color]color.Color {
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
