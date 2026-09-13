package logger

import (
	"io"
	"strings"
	"testing"
)

func TestColorsSetting(t *testing.T) {
	for name, tt := range map[string]struct {
		enable   bool
		colorful bool
	}{
		"enabled colors are rendered on non-tty output": {
			enable:   true,
			colorful: true,
		},
		"disabled colors are not rendered": {
			enable:   false,
			colorful: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			l := New(io.Discard)
			if tt.enable {
				l.EnableColors()
			} else {
				l.DisableColors()
			}

			painted := l.Paint(ColorRed, "text")
			if colorful := strings.Contains(painted, "\x1b["); colorful != tt.colorful {
				t.Errorf("expected colorful output to be %v, got: %q", tt.colorful, painted)
			}
		})
	}
}
