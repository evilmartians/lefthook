package log

import (
	"github.com/charmbracelet/lipgloss"
)

const skipLogPadding = 2

func Skip(name, reason string) {
	if !Settings.LogSkips() {
		return
	}

	Styled().
		WithLeftBorder(lipgloss.NormalBorder(), ColorCyan).
		WithPadding(skipLogPadding).
		Info(
			Cyan(Bold(name)) + " " +
				Gray("(skip)") + " " +
				Yellow(reason),
		)
}
