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

	"github.com/evilmartians/lefthook/internal/version"
)

var (
	ColorRed    lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#ff6347", ANSI256: "196", ANSI: "9"}
	ColorGreen  lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#32cd32", ANSI256: "148", ANSI: "2"}
	ColorYellow lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#fada5e", ANSI256: "191", ANSI: "11"}
	ColorCyan   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#70C0BA", ANSI256: "37", ANSI: "14"}
	GolorGray   lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#808080", ANSI256: "244", ANSI: "7"}
	colorBorder lipgloss.TerminalColor = lipgloss.Color("#383838")

	std = New()

	separatorWidth  = 36
	separatorMargin = 2
	padding         = 2
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

	ColorAuto = iota
	ColorOn
	ColorOff
)

type StyleLogger struct {
	style lipgloss.Style
}

type Logger struct {
	level   Level
	out     io.Writer
	mu      sync.Mutex
	colors  int
	names   []string
	spinner *spinner.Spinner
}

func New() *Logger {
	return &Logger{
		level:  InfoLevel,
		out:    os.Stdout,
		colors: ColorAuto,
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

func Colors() int {
	return std.colors
}

func Colorized() bool {
	return std.colors == ColorAuto || std.colors == ColorOn
}

func StartSpinner() {
	std.spinner.Start()
}

func StopSpinner() {
	std.spinner.Stop()
}

func Styled() StyleLogger {
	return StyleLogger{
		style: lipgloss.NewStyle(),
	}
}

func (s StyleLogger) WithLeftBorder(border lipgloss.Border, color lipgloss.TerminalColor) StyleLogger {
	s.style = s.style.BorderStyle(border).BorderLeft(true).BorderForeground(color)

	return s
}

func (s StyleLogger) WithPadding(m int) StyleLogger {
	s.style = s.style.PaddingLeft(m)

	return s
}

func (s StyleLogger) Info(str string) {
	Info(
		lipgloss.JoinVertical(
			lipgloss.Left,
			s.style.Render(str),
		),
	)
}

func Debug(args ...interface{}) {
	res := strings.TrimSpace(fmt.Sprint(args...))
	std.Debug(color(GolorGray).Render(res))
}

func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	std.Info(args...)
}

func InfoPad(s string) {
	Info(
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(ColorCyan).
			Render(s),
	)
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
	if colors == nil {
		return
	}

	switch typedColors := colors.(type) {
	case string:
		switch typedColors {
		case "on":
			std.colors = ColorOn
		case "off":
			std.colors = ColorOff
			setColor(lipgloss.NoColor{}, &ColorRed)
			setColor(lipgloss.NoColor{}, &ColorGreen)
			setColor(lipgloss.NoColor{}, &ColorYellow)
			setColor(lipgloss.NoColor{}, &ColorCyan)
			setColor(lipgloss.NoColor{}, &GolorGray)
			setColor(lipgloss.NoColor{}, &colorBorder)
		default:
			std.colors = ColorAuto
		}
	case bool:
		if typedColors {
			std.colors = ColorOn
			return
		}

		std.colors = ColorOff
		setColor(lipgloss.NoColor{}, &ColorRed)
		setColor(lipgloss.NoColor{}, &ColorGreen)
		setColor(lipgloss.NoColor{}, &ColorYellow)
		setColor(lipgloss.NoColor{}, &ColorCyan)
		setColor(lipgloss.NoColor{}, &GolorGray)
		setColor(lipgloss.NoColor{}, &colorBorder)
	case map[string]interface{}:
		std.colors = ColorOn
		setColor(typedColors["red"], &ColorRed)
		setColor(typedColors["green"], &ColorGreen)
		setColor(typedColors["yellow"], &ColorYellow)
		setColor(typedColors["cyan"], &ColorCyan)
		setColor(typedColors["gray"], &GolorGray)
		setColor(typedColors["gray"], &colorBorder)
	default:
		std.colors = ColorAuto
	}
}

func setColor(colorCode interface{}, adaptiveColor *lipgloss.TerminalColor) {
	var code string
	switch typedCode := colorCode.(type) {
	case int:
		code = strconv.Itoa(typedCode)
	case string:
		code = typedCode
	case lipgloss.NoColor:
		*adaptiveColor = typedCode
		return
	default:
		return
	}

	if len(code) == 0 {
		return
	}

	*adaptiveColor = lipgloss.Color(code)
}

func Cyan(s string) string {
	return color(ColorCyan).Render(s)
}

func Green(s string) string {
	return color(ColorGreen).Render(s)
}

func Red(s string) string {
	return color(ColorRed).Render(s)
}

func Yellow(s string) string {
	return color(ColorYellow).Render(s)
}

func Gray(s string) string {
	return color(GolorGray).Render(s)
}

func Bold(s string) string {
	if !Colorized() {
		return lipgloss.NewStyle().Render(s)
	}

	return lipgloss.NewStyle().Bold(true).Render(s)
}

func LogMeta(hookName string) {
	name := "🥊 lefthook "
	if !Colorized() {
		name = "lefthook "
	}

	box(
		Cyan(name)+Gray(fmt.Sprintf("v%s", version.Version(false))),
		Gray("hook: ")+Bold(hookName),
	)
}

func Success(indent int, name string, duration time.Duration) {
	format := "%s✔️ %s %s\n"
	if !Colorized() {
		format = "%s✓ %s %s\n"
	}
	Infof(
		format,
		strings.Repeat("  ", indent),
		Green(name),
		Gray(fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}

func Failure(indent int, name, failText string, duration time.Duration) {
	if len(failText) != 0 {
		failText = fmt.Sprintf(": %s", failText)
	}

	format := "%s🥊 %s%s %s\n"
	if !Colorized() {
		format = "%s✗ %s%s %s\n"
	}
	Infof(
		format,
		strings.Repeat("  ", indent),
		Red(name),
		Red(failText),
		Gray(fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}

func box(left, right string) {
	Info(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, false, true, true).
				BorderForeground(colorBorder).
				Padding(0, 1).
				Render(left),
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, true, false).
				BorderForeground(colorBorder).
				Padding(0, 1).
				Render(right),
		),
	)
}

func Separate(s string) {
	Info(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(colorBorder).
				Width(separatorWidth).
				MarginLeft(separatorMargin).
				Render(""),
			s,
		),
	)
}

func color(clr lipgloss.TerminalColor) lipgloss.Style {
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

func (l *Logger) Debug(args ...string) {
	leftBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(colorBorder).
		PaddingLeft(padding).
		Render(args...)
	l.Log(DebugLevel, leftBorder)
}

func (l *Logger) Error(args ...string) {
	leftBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(ColorRed).
		PaddingLeft(padding).
		Render(args...)
	l.Log(ErrorLevel, leftBorder)
}

func (l *Logger) Warn(args ...string) {
	leftBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(ColorYellow).
		PaddingLeft(padding).
		Render(args...)
	l.Log(WarnLevel, leftBorder)
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

	capacity := len(l.names)
	if capacity > 0 {
		capacity--
	}
	newNames := make([]string, 0, capacity)
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
