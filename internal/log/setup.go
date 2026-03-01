package log

import "github.com/charmbracelet/lipgloss"

func LogSetup() {
	Styled().
		WithLeftBorder(lipgloss.ThickBorder(), ColorYellow).
		WithPadding(execLogPadding).
		Info(Yellow("setup ❯ "))
}
