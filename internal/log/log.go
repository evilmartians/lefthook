package log

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorRed lipgloss.TerminalColor = lipgloss.CompleteAdaptiveColor{
		Dark:  lipgloss.CompleteColor{TrueColor: "#ff6347", ANSI256: "196", ANSI: "9"},
		Light: lipgloss.CompleteColor{TrueColor: "#d70000", ANSI256: "160", ANSI: "1"},
	}

	colorGreen lipgloss.TerminalColor = lipgloss.CompleteAdaptiveColor{
		Dark:  lipgloss.CompleteColor{TrueColor: "#76ff7a", ANSI256: "155", ANSI: "10"},
		Light: lipgloss.CompleteColor{TrueColor: "#afd700", ANSI256: "148", ANSI: "2"},
	}

	colorYellow lipgloss.TerminalColor = lipgloss.CompleteAdaptiveColor{
		Dark:  lipgloss.CompleteColor{TrueColor: "#fada5e", ANSI256: "191", ANSI: "11"},
		Light: lipgloss.CompleteColor{TrueColor: "#ffaf00", ANSI256: "214", ANSI: "3"},
	}

	colorCyan lipgloss.TerminalColor = lipgloss.CompleteAdaptiveColor{
		Dark:  lipgloss.CompleteColor{TrueColor: "#70C0BA", ANSI256: "37", ANSI: "14"},
		Light: lipgloss.CompleteColor{TrueColor: "#00af87", ANSI256: "36", ANSI: "6"},
	}

	colorGray lipgloss.TerminalColor = lipgloss.CompleteAdaptiveColor{
		Dark:  lipgloss.CompleteColor{TrueColor: "#808080", ANSI256: "244", ANSI: "7"},
		Light: lipgloss.CompleteColor{TrueColor: "#4e4e4e", ANSI256: "239", ANSI: "8"},
	}

	std = New()
)

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

func SetColors(colors interface{}) {
	switch typedColors := colors.(type) {
	case bool:
		std.colors = typedColors
		return
	case map[string]interface{}:
		std.colors = true
		setColor(typedColors["red"], &colorRed)
		setColor(typedColors["green"], &colorGreen)
		setColor(typedColors["yellow"], &colorYellow)
		setColor(typedColors["cyan"], &colorCyan)
		setColor(typedColors["gray"], &colorGray)
		return
	default:
		std.colors = true
	}
}

func setColor(colorCode interface{}, adaptiveColor *lipgloss.TerminalColor) {
	if colorCode == nil {
		return
	}

	var code string
	switch typedCode := colorCode.(type) {
	case int:
		code = strconv.Itoa(typedCode)
	case string:
		code = typedCode
	default:
		return
	}

	if len(code) == 0 {
		return
	}

	*adaptiveColor = lipgloss.Color(code)
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
	if !std.colors {
		return lipgloss.NewStyle().Render(s)
	}

	return lipgloss.NewStyle().Bold(true).Render(s)
}

func color(clr lipgloss.TerminalColor) lipgloss.Style {
	if !std.colors {
		return lipgloss.NewStyle()
	}

	return lipgloss.NewStyle().Foreground(clr)
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
