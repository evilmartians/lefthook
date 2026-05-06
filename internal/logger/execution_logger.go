package logger

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

const (
	separatorWidth  = 36
	separatorMargin = 2
	skipLogPadding  = 2
)

var colorBorder color.Color = lipgloss.Color("#383838")

type ExecutionLogger struct {
	*Logger

	settings *ExecutionSettings
	Spinner  *Spinner
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
					l.Warnf("Unknown output setting: %v", option)
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
		Spinner:  l.Spinner,
	}
}

func (el *ExecutionLogger) Enabled(setting ExecutionSetting) bool {
	return el.settings.enabled(setting)
}

func (el *ExecutionLogger) LogSkipped(name, reason string) {
	if !el.Enabled(LogSkips) {
		return
	}

	el.Info(
		lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(el.Logger.colors.get(ColorCyan)).
			PaddingLeft(skipLogPadding).
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
