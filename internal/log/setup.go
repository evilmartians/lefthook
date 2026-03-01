package log

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

func LogSetup(r io.Reader) {
	go func() {
		if !Settings.LogSetup() {
			_, _ = io.Copy(io.Discard, r)
			return
		}

		Styled().
			WithLeftBorder(lipgloss.ThickBorder(), ColorYellow).
			WithPadding(execLogPadding).
			Info(Yellow("setup ❯ "))

		_, _ = io.Copy(os.Stdout, r)
	}()
}
