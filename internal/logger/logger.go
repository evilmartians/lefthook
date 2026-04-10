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

type Color uint8

const (
	ColorCyan Color = iota
	ColorGray
	ColorGreen
	ColorRed
	ColorYellow
	ColorWhite
)

var profile = colorprofile.Detect(os.Stdout, os.Environ())
var complete = lipgloss.Complete(profile)
var DefaultColors map[Color]color.Color = map[Color]color.Color{
	ColorCyan:   complete(lipgloss.Color("37"), lipgloss.Color("14"), lipgloss.Color("#70C0BA")),
	ColorGray:   complete(lipgloss.Color("7"), lipgloss.Color("244"), lipgloss.Color("#808080")),
	ColorGreen:  complete(lipgloss.Color("2"), lipgloss.Color("148"), lipgloss.Color("#32cd32")),
	ColorRed:    complete(lipgloss.Color("9"), lipgloss.Color("196"), lipgloss.Color("#ff6347")),
	ColorYellow: complete(lipgloss.Color("11"), lipgloss.Color("191"), lipgloss.Color("#fada5e")),
}
var NoColors map[Color]color.Color = map[Color]color.Color{
	ColorCyan:   lipgloss.NoColor{},
	ColorGray:   lipgloss.NoColor{},
	ColorGreen:  lipgloss.NoColor{},
	ColorRed:    lipgloss.NoColor{},
	ColorYellow: lipgloss.NoColor{},
}
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
	l.Print(message)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...string) {
	if l.level < LevelWarn {
		return
	}

	message := border.BorderForeground(l.colors[ColorYellow]).Render(args...)
	l.Print(message)
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.level < LevelWarn {
		return
	}

	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...any) {
	if l.level < LevelInfo {
		return
	}

	l.Print(args...)
}

func (l *Logger) Infof(format string, args ...any) {
	if l.level < LevelInfo {
		return
	}

	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(args ...string) {
	if l.level < LevelDebug {
		return
	}

	message := border.BorderForeground(l.colors[ColorGray]).Render(args...)
	l.Print(message)
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.level < LevelDebug {
		return
	}

	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Print(args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Spinner.active() {
		l.Spinner.Stop()
		defer l.Spinner.Start()
	}

	_, _ = fmt.Fprintln(l.out, args...)
}

// func (l *Logger) Printf(format string, args ...any) {
// 	l.Print(fmt.Sprintf(format, args...))
// }

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
