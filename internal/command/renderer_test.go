package command

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestConfigureRenderer(t *testing.T) {
	t.Run("non-TTY stdin forces ANSI profile", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		defer w.Close()

		// Reset the default renderer to a known state before testing.
		lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stdout))

		configureRenderer(r)

		got := lipgloss.DefaultRenderer().ColorProfile()
		if got != termenv.ANSI {
			t.Errorf("expected ANSI profile for non-TTY stdin, got %v", got)
		}
	})
}
