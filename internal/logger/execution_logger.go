package logger

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/evilmartians/lefthook/v2/internal/version"
)

const (
	separatorWidth  = 36
	separatorMargin = 2
	padding         = 2
)

var colorBorder color.Color = lipgloss.Color("#383838")

// ExecutionLogger wraps the Logger but also provides additional methods for logging execution
// with predefined style to simplify the caller side.
type ExecutionLogger struct {
	// Inherit all Logger methods
	*Logger

	// Control printing of execution sections
	settings *ExecutionSettings
}

func (l *Logger) NewExecutionLogger(configs ...any) *ExecutionLogger {
	settings := NewExecutionSettings()

	for _, config := range configs {
		switch c := config.(type) {
		case bool:
			if c {
				settings.enable(executionFull)
			}
		case []any:
			for _, option := range c {
				name, ok := option.(string)
				if !ok {
					l.Warnf("Unknown output setting: %#v", option)
					continue
				}
				if len(name) == 0 {
					continue
				}

				setting, err := nameToSetting(name)
				if err != nil {
					l.Warn(err)
					continue
				}
				settings.enable(setting)
			}
		case string:
			names := strings.Split(c, ",")
			for _, name := range names {
				if len(name) == 0 {
					continue
				}

				setting, err := nameToSetting(name)
				if err != nil {
					l.Warn(err)
					break // fallthrough to default
				}
				settings.enable(setting)
			}
		default:
			settings.enable(executionFull)
		}
	}

	return &ExecutionLogger{
		Logger:   l,
		settings: settings,
	}
}

func (el *ExecutionLogger) Enabled(setting ExecutionSetting) bool {
	return el.settings.enabled(setting)
}

func (el *ExecutionLogger) LogSkipped(name, reason string) {
	if !el.Enabled(LogSkips) {
		return
	}

	el.Logger.Info(
		lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(el.Logger.colors.get(ColorCyan)).
			PaddingLeft(padding).
			Render(
				el.Logger.Paint(ColorCyan, lipgloss.NewStyle().Bold(true).Render(name)) + " " +
					el.Logger.Paint(ColorGray, "(skip)") + " " +
					el.Logger.Paint(ColorYellow, reason),
			),
	)
}

func (el *ExecutionLogger) LogSeparator() {
	el.Logger.log(LevelInfo,
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder).
			Width(separatorWidth).
			MarginLeft(separatorMargin).
			Render(""),
	)
}

func (el *ExecutionLogger) LogExecution(name string, err error, out io.Reader) {
	if err == nil && !el.Enabled(LogExecution) {
		return
	}

	var execLog string
	var color Color
	switch {
	case !el.Enabled(LogExecutionInfo):
		execLog = ""
	case err != nil:
		execLog = el.Paint(ColorRed, fmt.Sprintf("%s ❯ ", name))
		color = ColorRed
	default:
		execLog = el.Paint(ColorCyan, fmt.Sprintf("%s ❯ ", name))
		color = ColorCyan
	}

	if execLog != "" {
		el.Logger.Info(
			lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderLeft(true).
				BorderForeground(el.Logger.colors.get(color)).
				PaddingLeft(padding).
				Render(execLog),
		)
	}

	if err == nil && !el.Enabled(LogExecutionOutput) {
		return
	}

	if out != nil {
		el.Logger.Info(out)
	}

	if err != nil {
		el.Logger.Infof("%s", err)
	}
}

func (el *ExecutionLogger) LogSetup(r io.Reader) {
	go func() {
		if !el.Enabled(LogSetup) {
			_, _ = io.Copy(io.Discard, r)
			return
		}

		el.Logger.Info(
			lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderLeft(true).
				BorderForeground(el.Logger.colors.get(ColorYellow)).
				Padding(padding).
				Render(el.Logger.Paint(ColorYellow, "setup ❯ ")),
		)

		_, _ = io.Copy(os.Stdout, r)
	}()
}

func (el *ExecutionLogger) LogMeta(hookName string) {
	name := "🥊 lefthook "
	if el.Logger.NoColors() {
		name = "lefthook "
	}

	el.Logger.Info(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, false, true, true).
				BorderForeground(colorBorder).
				Padding(0, 1).
				Render(
					el.Logger.Paint(ColorCyan, name),
					el.Logger.Paint(ColorGray, fmt.Sprintf("v%s", version.Version(false))),
				),
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, true, false).
				BorderForeground(colorBorder).
				Padding(0, 1).
				Render(
					el.Logger.Paint(ColorGray, "hook: "),
					lipgloss.NewStyle().Bold(true).Render(hookName),
				),
		),
	)
}

func (el *ExecutionLogger) LogSuccess(indent int, name string, duration time.Duration) {
	var format string
	if el.Logger.NoColors() {
		format = "%s✓ %s %s"
	} else {
		format = "%s✔️ %s %s"
	}

	el.Logger.Infof(
		format,
		strings.Repeat("  ", indent),
		el.Logger.Paint(ColorGreen, name),
		el.Logger.Paint(ColorGray, fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}

func (el *ExecutionLogger) LogFailure(indent int, name, failText string, duration time.Duration) {
	if len(failText) != 0 {
		failText = fmt.Sprintf(": %s", failText)
	}

	var format string
	if el.Logger.NoColors() {
		format = "%s✗ %s%s %s"
	} else {
		format = "%s🥊 %s%s %s"
	}

	el.Logger.Infof(
		format,
		strings.Repeat("  ", indent),
		el.Logger.Paint(ColorRed, name),
		el.Logger.Paint(ColorRed, failText),
		el.Logger.Paint(ColorGray, fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}
