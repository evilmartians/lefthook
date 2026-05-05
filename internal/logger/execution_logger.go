package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	separatorWidth  = 36
	separatorMargin = 2
	skipLogPadding  = 2
)

var colorBorder lipgloss.TerminalColor = lipgloss.Color("#383838")

type ExecutionLogger struct {
	logger   *Logger
	settings *ExecutionSettings
	Spinner  *Spinner
}

func (l *logger) NewExecutionLogger(configs ...any) *ExecutionLogger {
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
					logger.Warnf("Unknown output setting: %v", option)
					continue
				}

				setting, err := nameToSetting(name)
				if err != nil {
					logger.Warn(err)
					continue
				}
				settings.enable(setting)
			}
		case string:
			names := strings.Split(name, ",")
			for _, name := range names {
				setting, err := nameToSetting(name)
				if err != nil {
					logger.Warn(err)
					fallthrough
				}
				settings.enable(setting)
			}
		default:
			settings.enable(executionFull)
		}
	}

	return &ExecutionLogger{
		logger:   logger,
		settings: settings,
		Spinner:  logger.Spinner,
	}
}

func (el *ExecutionLogger) Enabled(setting ExecutionSetting) bool {
	return el.settings.enabled(setting)
}

func (el *ExecutionLogger) Debug(args ...any) {
	el.logger.Debug(args...)
}

func (el *ExecutionLogger) Info(args ...any) {
	el.logger.Info(args...)
}

func (el *ExecutionLogger) Infof(format string, args ...any) {
	el.logger.Infof(format, args...)
}

func (el *ExecutionLogger) Warn(args ...any) {
	el.logger.Warn(args...)
}

func (el *ExecutionLogger) Warn(format string, args ...any) {
	el.logger.Warnf(format, args...)
}

func (el *ExecutionLogger) log(level Level, args ...any) {
	el.logger.log(level, args...)
}

func (el *ExecutionLogger) NoColors() bool {
	return el.logger.colors == NoColors
}

func (el *ExecutionLogger) LogSkipped(name, reason string) {
	if !el.Enabled(LogSkips) {
		return
	}

	el.Info(
		lipgloss.NewStyle().
			WithLeftBorder(lipgloss.NormalBorder(), ColorCyan).
			WithPadding(skipLogPadding).
			el.logger.Paint(ColorCyan, Bold(name)) + " " +
			el.logger.Paint(ColorGray, "(skip)") + " " +
			el.logger.Paint(ColorYellow, reason),
	)
}

func (el *ExecutionLogger) LogSeparator() {
	el.logger.log(LevelInfo,
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
	if el.logger.colors == NoColors {
		format = "%s✓ %s %s"
	} else {
		format = "%s✔️ %s %s"
	}

	el.logger.Infof(
		format,
		strings.Repeat("  ", indent),
		el.logger.Paint(ColorGreen, name),
		el.logger.Paint(ColorGray, fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}

func (el *ExecutionLogger) LogFailure(indent int, name, failText string, duration time.Duration) {
	if len(failText) != 0 {
		failText = fmt.Sprintf(": %s", failText)
	}

	var format string
	if el.logger.colors == NoColors {
		format = "%s✗ %s%s %s"
	} else {
		format = "%s🥊 %s%s %s"
	}

	el.logger.Infof(
		format,
		strings.Repeat("  ", indent),
		el.logger.Paint(ColorRed, name),
		el.logger.Paint(ColorRed, failText),
		el.logger.Paint(ColorGray, fmt.Sprintf("(%.2f seconds)", duration.Seconds())),
	)
}
