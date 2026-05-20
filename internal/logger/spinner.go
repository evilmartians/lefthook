package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
)

const (
	spinnerCharSet     = 14
	spinnerRefreshRate = 100 * time.Millisecond
	spinnerText        = " waiting"
)

type Spinner struct {
	mu            sync.Mutex
	terminalWidth int
	spinner       *spinner.Spinner
	names         []string
}

func NewSpinner() *Spinner {
	return &Spinner{
		names:         make([]string, 0, 10), //nolint:mnd // reduce extra allocations
		terminalWidth: terminalWidth(),
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

func (s *Spinner) Start() {
	s.spinner.Start()
}

func (s *Spinner) Stop() {
	s.spinner.Stop()
}

func (s *Spinner) AddName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.spinner.Active() {
		s.spinner.Stop()
		defer s.spinner.Start()
	}

	s.names = append(s.names, name)
	s.spinner.Suffix = formatSpinnerSuffix(s.names, s.terminalWidth)
}

func (s *Spinner) RemoveName(nameToRemove string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.spinner.Active() {
		s.spinner.Stop()
		defer s.spinner.Start()
	}

	j := 0
	for _, name := range s.names {
		if name == nameToRemove {
			continue
		}

		s.names[j] = name
		j++
	}

	s.names = s.names[:j]
	s.spinner.Suffix = formatSpinnerSuffix(s.names, s.terminalWidth)
}

func formatSpinnerSuffix(names []string, width int) string {
	if len(names) == 0 {
		return spinnerText
	}

	if width <= 0 {
		return fmt.Sprintf("%s: %s", spinnerText, strings.Join(names, ", "))
	}

	// Width calculation: Reserve space for spinner character (1) + space (1) + padding (8)
	// This accounts for the spinning character and reasonable display margin
	const spinnerReservedWidth = 10
	availableWidth := width - spinnerReservedWidth

	// Strategy 1: Try to fit all names with full formatting
	fullSuffix := fmt.Sprintf("%s: %s", spinnerText, strings.Join(names, ", "))
	if runewidth.StringWidth(fullSuffix) <= availableWidth {
		return fullSuffix
	}

	// Strategy 2: Try showing just the count
	countSuffix := fmt.Sprintf("%s: %d hook%s", spinnerText, len(names), pluralize(len(names)))
	if runewidth.StringWidth(countSuffix) <= availableWidth {
		return countSuffix
	}

	// Strategy 3: Show as many individual names as possible
	return formatWithPartialNames(names, availableWidth)
}

// terminalWidth attempts to detect the current terminal width.
func terminalWidth() int {
	// Check if we're writing to a TTY
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return 0 // Not a terminal, don't constrain
	}

	// Try to get terminal size
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		return 0 // Can't determine size, don't constrain
	}

	return width
}

// formatWithPartialNames shows as many hook names as possible, then adds count for remaining.
func formatWithPartialNames(names []string, availableWidth int) string {
	if len(names) == 0 {
		return spinnerText
	}

	baseText := spinnerText + ": "
	baseWidth := runewidth.StringWidth(baseText)
	remainingWidth := availableWidth - baseWidth

	// Try to fit names one by one
	var fittingNames []string
	currentWidth := 0

	for i, name := range names {
		nameWidth := runewidth.StringWidth(name)

		// Add comma and space for all but first name
		if i > 0 {
			nameWidth += 2 // ", "
		}

		// Check if we need space for "... (N more)" suffix
		remainingCount := len(names) - i
		if remainingCount > 1 {
			moreSuffix := fmt.Sprintf(", ... (%d more)", remainingCount-1)
			moreSuffixWidth := runewidth.StringWidth(moreSuffix)

			if currentWidth+nameWidth+moreSuffixWidth > remainingWidth {
				// Add the "more" suffix and break
				if len(fittingNames) > 0 {
					return fmt.Sprintf("%s%s, ... (%d more)", baseText, strings.Join(fittingNames, ", "), remainingCount)
				}
				// If we can't fit even one name, just show count
				return fmt.Sprintf("%s%d hook%s", baseText, len(names), pluralize(len(names)))
			}
		}

		if currentWidth+nameWidth <= remainingWidth {
			fittingNames = append(fittingNames, name)
			currentWidth += nameWidth
		} else {
			// This name doesn't fit
			if len(fittingNames) == 0 {
				// Can't fit any names, just show count
				return fmt.Sprintf("%s%d hook%s", baseText, len(names), pluralize(len(names)))
			}
			// Show what we have plus count
			remainingCount := len(names) - len(fittingNames)
			return fmt.Sprintf("%s%s, ... (%d more)", baseText, strings.Join(fittingNames, ", "), remainingCount)
		}
	}

	// All names fit
	return fmt.Sprintf("%s%s", baseText, strings.Join(fittingNames, ", "))
}

// pluralize returns "s" for counts != 1, empty string otherwise.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
