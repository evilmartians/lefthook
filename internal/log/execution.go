package log

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

const execLogPadding = 2

func Execution(name string, err error, out io.Reader) {
	if err == nil && !Settings.LogExecution() {
		return
	}

	var execLog string
	var color lipgloss.TerminalColor
	switch {
	case !Settings.LogExecutionInfo():
		execLog = ""
	case err != nil:
		execLog = Red(fmt.Sprintf("%s ❯ ", name))
		color = ColorRed
	default:
		execLog = Cyan(fmt.Sprintf("%s ❯ ", name))
		color = ColorCyan
	}

	if execLog != "" {
		Styled().
			WithLeftBorder(lipgloss.ThickBorder(), color).
			WithPadding(execLogPadding).
			Info(execLog)
		Info()
	}

	if err == nil && !Settings.LogExecutionOutput() {
		return
	}

	if out != nil {
		Info(out)
	}

	if err != nil {
		Infof("%s", err)
	}
}
