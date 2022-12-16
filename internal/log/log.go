package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorCyan   = "#70C0BA"
	colorYellow = "#fada5e"
	colorRed    = "#ff6347"
	colorGreen  = "#76ff7a"
	colorGray   = "#808080"
)

var std = New()

type Level uint32

const (
	ErrorLevel Level = iota
	WarnLevel
	InfoLevel
	DebugLevel

	spinnerCharSet     = 14
	spinnerRefreshRate = 100 * time.Millisecond
	spinnerText        = " waiting"
)

type Logger struct {
	level   Level
	out     io.Writer
	mu      sync.Mutex
	colors  bool
	names   []string
	spinner *spinner.Spinner
}

func New() *Logger {
	return &Logger{
		level:  InfoLevel,
		out:    os.Stdout,
		colors: true,
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

func StartSpinner() {
	std.spinner.Start()
}

func StopSpinner() {
	std.spinner.Stop()
}

func Debug(args ...interface{}) {
	res := fmt.Sprint(args...)
	std.Debug(color(colorGray).Render(res))
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
	std.Warn(Yellow(res))
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
	std.colors = enable
}

func Cyan(s string) string {
	return color(colorCyan).Render(s)
}

func Green(s string) string {
	return color(colorGreen).Render(s)
}

func Red(s string) string {
	return color(colorRed).Render(s)
}

func Yellow(s string) string {
	return color(colorYellow).Render(s)
}

func Gray(s string) string {
	return color(colorGray).Render(s)
}

func Bold(s string) string {
	return lipgloss.NewStyle().Bold(true).Render(s)
}

func color(colorCode string) lipgloss.Style {
	if !std.colors {
		return lipgloss.NewStyle()
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode))
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

func SetName(name string) {
	std.SetName(name)
}

func UnsetName(name string) {
	std.UnsetName(name)
}

func (l *Logger) SetName(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner.Active() {
		l.spinner.Stop()
		defer l.spinner.Start()
	}

	l.names = append(l.names, name)
	l.spinner.Suffix = fmt.Sprintf("%s: %s", spinnerText, strings.Join(l.names, ", "))
}

func (l *Logger) UnsetName(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner.Active() {
		l.spinner.Stop()
		defer l.spinner.Start()
	}

	newNames := make([]string, 0, len(l.names)-1)
	for _, n := range l.names {
		if n != name {
			newNames = append(newNames, n)
		}
	}

	l.names = newNames

	if len(l.names) != 0 {
		l.spinner.Suffix = fmt.Sprintf("%s: %s", spinnerText, strings.Join(l.names, ", "))
	} else {
		l.spinner.Suffix = spinnerText
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

	if l.spinner.Active() {
		l.spinner.Stop()
		defer l.spinner.Start()
	}

	_, _ = fmt.Fprintln(l.out, args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner.Active() {
		l.spinner.Stop()
		defer l.spinner.Start()
	}

	_, _ = fmt.Fprintf(l.out, format, args...)
}

func (l *Logger) IsLevelEnabled(level Level) bool {
	return l.level >= level
}
