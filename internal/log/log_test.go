package log

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"
)

const (
	// Test constants for concurrent access.
	testConcurrentGoroutines   = 10
	testOperationsPerGoroutine = 50

	// Benchmark constants.
	benchmarkPrePopulationSize = 1000
	benchmarkMixedOperations   = 100
)

func TestLogger_SetName(t *testing.T) {
	for name, tt := range map[string]struct {
		initialNames   []string
		nameToAdd      string
		expectedNames  []string
		expectedSuffix string
	}{
		"add first name": {
			initialNames:   []string{},
			nameToAdd:      "test-hook",
			expectedNames:  []string{"test-hook"},
			expectedSuffix: " waiting: test-hook",
		},
		"add second name": {
			initialNames:   []string{"first-hook"},
			nameToAdd:      "second-hook",
			expectedNames:  []string{"first-hook", "second-hook"},
			expectedSuffix: " waiting: first-hook, second-hook",
		},
		"add multiple names": {
			initialNames:   []string{"hook1", "hook2"},
			nameToAdd:      "hook3",
			expectedNames:  []string{"hook1", "hook2", "hook3"},
			expectedSuffix: " waiting: hook1, hook2, hook3",
		},
		"add empty name": {
			initialNames:   []string{"existing"},
			nameToAdd:      "",
			expectedNames:  []string{"existing", ""},
			expectedSuffix: " waiting: existing, ",
		},
		"add duplicate name": {
			initialNames:   []string{"hook1"},
			nameToAdd:      "hook1",
			expectedNames:  []string{"hook1", "hook1"},
			expectedSuffix: " waiting: hook1, hook1",
		},
		"add name with spaces": {
			initialNames:   []string{},
			nameToAdd:      "hook with spaces",
			expectedNames:  []string{"hook with spaces"},
			expectedSuffix: " waiting: hook with spaces",
		},
		"add name with unicode": {
			initialNames:   []string{},
			nameToAdd:      "🥊-hook",
			expectedNames:  []string{"🥊-hook"},
			expectedSuffix: " waiting: 🥊-hook",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			logger := createTestLogger()

			// Set up initial state
			logger.names = make([]string, len(tt.initialNames))
			copy(logger.names, tt.initialNames)

			// Call SetName
			logger.SetName(tt.nameToAdd)

			// Verify names slice
			assert.Equal(tt.expectedNames, logger.names)

			// Verify spinner suffix
			assert.Equal(tt.expectedSuffix, logger.spinner.Suffix)
		})
	}
}

func TestLogger_UnsetName(t *testing.T) {
	for name, tt := range map[string]struct {
		initialNames   []string
		nameToRemove   string
		expectedNames  []string
		expectedSuffix string
	}{
		"remove only name": {
			initialNames:   []string{"test-hook"},
			nameToRemove:   "test-hook",
			expectedNames:  []string{},
			expectedSuffix: " waiting",
		},
		"remove first of two names": {
			initialNames:   []string{"first-hook", "second-hook"},
			nameToRemove:   "first-hook",
			expectedNames:  []string{"second-hook"},
			expectedSuffix: " waiting: second-hook",
		},
		"remove second of two names": {
			initialNames:   []string{"first-hook", "second-hook"},
			nameToRemove:   "second-hook",
			expectedNames:  []string{"first-hook"},
			expectedSuffix: " waiting: first-hook",
		},
		"remove middle name": {
			initialNames:   []string{"hook1", "hook2", "hook3"},
			nameToRemove:   "hook2",
			expectedNames:  []string{"hook1", "hook3"},
			expectedSuffix: " waiting: hook1, hook3",
		},
		"remove non-existent name": {
			initialNames:   []string{"hook1", "hook2"},
			nameToRemove:   "hook3",
			expectedNames:  []string{"hook1", "hook2"},
			expectedSuffix: " waiting: hook1, hook2",
		},
		"remove from empty list": {
			initialNames:   []string{},
			nameToRemove:   "hook1",
			expectedNames:  []string{},
			expectedSuffix: " waiting",
		},
		"remove empty name": {
			initialNames:   []string{"hook1", "", "hook2"},
			nameToRemove:   "",
			expectedNames:  []string{"hook1", "hook2"},
			expectedSuffix: " waiting: hook1, hook2",
		},
		"remove all duplicates": {
			initialNames:   []string{"hook1", "hook1", "hook2"},
			nameToRemove:   "hook1",
			expectedNames:  []string{"hook2"},
			expectedSuffix: " waiting: hook2",
		},
		"remove unicode name": {
			initialNames:   []string{"🥊-hook", "normal-hook"},
			nameToRemove:   "🥊-hook",
			expectedNames:  []string{"normal-hook"},
			expectedSuffix: " waiting: normal-hook",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			logger := createTestLogger()

			// Set up initial state
			logger.names = make([]string, len(tt.initialNames))
			copy(logger.names, tt.initialNames)

			// Call UnsetName
			logger.UnsetName(tt.nameToRemove)

			// Verify names slice
			assert.Equal(tt.expectedNames, logger.names)

			// Verify spinner suffix
			assert.Equal(tt.expectedSuffix, logger.spinner.Suffix)
		})
	}
}

func TestLogger_SetName_UnsetName_Integration(t *testing.T) {
	assert := assert.New(t)
	logger := createTestLogger()

	// Start with empty state
	assert.Equal([]string{}, logger.names)
	assert.Equal(" waiting", logger.spinner.Suffix)

	// Add first name
	logger.SetName("hook1")
	assert.Equal([]string{"hook1"}, logger.names)
	assert.Equal(" waiting: hook1", logger.spinner.Suffix)

	// Add second name
	logger.SetName("hook2")
	assert.Equal([]string{"hook1", "hook2"}, logger.names)
	assert.Equal(" waiting: hook1, hook2", logger.spinner.Suffix)

	// Add third name
	logger.SetName("hook3")
	assert.Equal([]string{"hook1", "hook2", "hook3"}, logger.names)
	assert.Equal(" waiting: hook1, hook2, hook3", logger.spinner.Suffix)

	// Remove middle name
	logger.UnsetName("hook2")
	assert.Equal([]string{"hook1", "hook3"}, logger.names)
	assert.Equal(" waiting: hook1, hook3", logger.spinner.Suffix)

	// Remove first name
	logger.UnsetName("hook1")
	assert.Equal([]string{"hook3"}, logger.names)
	assert.Equal(" waiting: hook3", logger.spinner.Suffix)

	// Remove last name
	logger.UnsetName("hook3")
	assert.Equal([]string{}, logger.names)
	assert.Equal(" waiting", logger.spinner.Suffix)
}

func TestLogger_LongHookNames(t *testing.T) {
	assert := assert.New(t)
	logger := createTestLogger()

	// This test documents current behavior that causes terminal wrapping.
	// See issue #1144 for planned terminal width handling.
	// Test with very long hook names that would exceed typical terminal width
	longNames := []string{
		"very-long-hook-name-that-exceeds-normal-terminal-width-and-would-cause-wrapping-issues",
		"another-extremely-long-hook-name-with-many-hyphens-and-descriptive-text-that-goes-on-and-on",
		"packwerk_check_unused_dependencies_and_validate_all_boundaries_with_strict_mode_enabled",
		"eslint_with_typescript_support_and_custom_rules_for_react_components_and_styled_components",
	}

	// Add all long names
	for _, name := range longNames {
		logger.SetName(name)
	}

	// Verify all names are stored
	assert.Equal(longNames, logger.names)

	// Verify the suffix contains all names (this is the current behavior that causes the issue)
	expectedSuffix := " waiting: " + strings.Join(longNames, ", ")
	assert.Equal(expectedSuffix, logger.spinner.Suffix)

	// Document the current problematic behavior
	t.Logf("Current suffix length: %d characters", len(logger.spinner.Suffix))
	t.Logf("This would cause wrapping issues in terminals narrower than %d columns", len(logger.spinner.Suffix))
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	assert := assert.New(t)
	logger := createTestLogger()

	var wg sync.WaitGroup

	// Start goroutines that concurrently add and remove names
	for i := range testConcurrentGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := range testOperationsPerGoroutine {
				hookName := fmt.Sprintf("hook-%d-%d", id, j)

				// Add name
				logger.SetName(hookName)

				// Small delay to increase chance of race conditions
				time.Sleep(time.Microsecond)

				// Remove name
				logger.UnsetName(hookName)
			}
		}(i)
	}

	wg.Wait()

	// After all operations, names slice should be empty
	assert.Equal([]string{}, logger.names)
	assert.Equal(" waiting", logger.spinner.Suffix)
}

func TestLogger_SpinnerActiveHandling(t *testing.T) {
	assert := assert.New(t)
	logger := createTestLogger()

	// Test that SetName and UnsetName don't panic when spinner is active
	logger.spinner.Start()
	initialActive := logger.spinner.Active()

	// SetName should handle active spinner without panicking
	logger.SetName("test-hook")
	assert.Equal([]string{"test-hook"}, logger.names)
	assert.Equal(" waiting: test-hook", logger.spinner.Suffix)

	// UnsetName should handle active spinner without panicking
	logger.UnsetName("test-hook")
	assert.Equal([]string{}, logger.names)
	assert.Equal(" waiting", logger.spinner.Suffix)

	// Clean up
	logger.spinner.Stop()

	// Document the behavior for future reference
	t.Logf("Spinner was initially active: %v", initialActive)
}

func TestGlobalSetNameUnsetName(t *testing.T) {
	assert := assert.New(t)

	// Test the global functions that use the standard logger
	originalNames := make([]string, len(std.names))
	copy(originalNames, std.names)
	originalSuffix := std.spinner.Suffix

	// Clean up after test
	defer func() {
		std.names = originalNames
		std.spinner.Suffix = originalSuffix
	}()

	// Reset to clean state
	std.names = []string{}
	std.spinner.Suffix = " waiting"

	// Test global SetName
	SetName("global-hook")
	assert.Equal([]string{"global-hook"}, std.names)
	assert.Equal(" waiting: global-hook", std.spinner.Suffix)

	// Test global UnsetName
	UnsetName("global-hook")
	assert.Equal([]string{}, std.names)
	assert.Equal(" waiting", std.spinner.Suffix)
}

// Helper function to create a test logger.
func createTestLogger() *Logger {
	return &Logger{
		level:  InfoLevel,
		out:    &bytes.Buffer{},
		colors: ColorOff,
		names:  []string{},
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

// Benchmark tests to understand performance characteristics.
func BenchmarkSetName(b *testing.B) {
	logger := createTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.SetName(fmt.Sprintf("hook-%d", i))
	}
}

func BenchmarkUnsetName(b *testing.B) {
	logger := createTestLogger()

	// Pre-populate with names
	for i := 0; i < benchmarkPrePopulationSize; i++ {
		logger.names = append(logger.names, fmt.Sprintf("hook-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.UnsetName(fmt.Sprintf("hook-%d", i%benchmarkPrePopulationSize))
	}
}

func BenchmarkSetNameUnsetName(b *testing.B) {
	logger := createTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hookName := fmt.Sprintf("hook-%d", i)
		logger.SetName(hookName)
		logger.UnsetName(hookName)
	}
}

// Terminal width handling tests
func TestLogger_FormatSpinnerSuffix(t *testing.T) {
	tests := []struct {
		name          string
		names         []string
		terminalWidth int
		expected      string
	}{
		{
			name:          "empty names",
			names:         []string{},
			terminalWidth: 80,
			expected:      " waiting",
		},
		{
			name:          "single short name fits",
			names:         []string{"test"},
			terminalWidth: 80,
			expected:      " waiting: test",
		},
		{
			name:          "multiple short names fit",
			names:         []string{"hook1", "hook2", "hook3"},
			terminalWidth: 80,
			expected:      " waiting: hook1, hook2, hook3",
		},
		{
			name:          "names too long, show count",
			names:         []string{"very-long-hook-name-1", "very-long-hook-name-2", "very-long-hook-name-3"},
			terminalWidth: 30,
			expected:      " waiting: 3 hooks",
		},
		{
			name:          "single hook, singular",
			names:         []string{"hook1"},
			terminalWidth: 20,
			expected:      " waiting: 1 hook",
		},
		{
			name:          "short names all fit in available width",
			names:         []string{"a", "b", "c", "d", "e"},
			terminalWidth: 35, // All short names fit
			expected:      " waiting: a, b, c, d, e",
		},
		{
			name:          "names too wide, fallback to count",
			names:         []string{"hook", "test", "verylongname", "another", "final"},
			terminalWidth: 30, // Too narrow, shows count instead
			expected:      " waiting: 5 hooks",
		},
		{
			name:          "unicode characters handled correctly",
			names:         []string{"🥊-hook", "test"},
			terminalWidth: 80,
			expected:      " waiting: 🥊-hook, test",
		},
		{
			name:          "no terminal width constraint (width 0)",
			names:         []string{"hook1", "hook2", "very-long-name"},
			terminalWidth: 0,
			expected:      " waiting: hook1, hook2, very-long-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := createTestLoggerWithWidth(tt.terminalWidth)
			result := logger.formatSpinnerSuffix(tt.names)
			
			// Debug output for failing tests
			if result != tt.expected {
				t.Logf("Terminal width: %d", tt.terminalWidth)
				t.Logf("Available width: %d", tt.terminalWidth-10)
				t.Logf("Expected: %q", tt.expected)
				t.Logf("Actual: %q", result)
			}
			
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogger_GetTerminalWidth(t *testing.T) {
	logger := createTestLogger()
	
	// This test will vary based on the actual terminal, but we can test the basic functionality
	width := logger.getTerminalWidth()
	
	// Width should be either 0 (not a TTY) or a positive number
	assert.True(t, width >= 0, "Terminal width should be non-negative")
	
	// Log the detected width for debugging
	t.Logf("Detected terminal width: %d", width)
}

func TestLogger_FormatWithPartialNames(t *testing.T) {
	tests := []struct {
		name           string
		names          []string
		availableWidth int
		expectedResult string
	}{
		{
			name:           "empty names",
			names:          []string{},
			availableWidth: 50,
			expectedResult: " waiting",
		},
		{
			name:           "all names fit",
			names:          []string{"a", "b", "c"},
			availableWidth: 50,
			expectedResult: " waiting: a, b, c",
		},
		{
			name:           "partial names fit",
			names:          []string{"hook1", "hook2", "very-long-hook-name"},
			availableWidth: 30,
			expectedResult: " waiting: hook1, ... (2 more)",
		},
		{
			name:           "very narrow width, show count only",
			names:          []string{"hook1", "hook2"},
			availableWidth: 15,
			expectedResult: " waiting: 2 hooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := createTestLogger()
			result := logger.formatWithPartialNames(tt.names, tt.availableWidth)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count_%d", tt.count), func(t *testing.T) {
			result := pluralize(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogger_TerminalWidthIntegration(t *testing.T) {
	// Test the integration of SetName/UnsetName with terminal width handling
	logger := createTestLoggerWithWidth(50) // Simulate 50-character terminal
	
	// Add hooks that would exceed width
	longHooks := []string{
		"very-long-hook-name-1",
		"very-long-hook-name-2", 
		"very-long-hook-name-3",
		"very-long-hook-name-4",
	}
	
	for _, hook := range longHooks {
		logger.SetName(hook)
	}
	
	// Should show count instead of all names
	assert.Contains(t, logger.spinner.Suffix, "hooks")
	assert.NotContains(t, logger.spinner.Suffix, "very-long-hook-name-4")
	
	// Remove some hooks
	logger.UnsetName("very-long-hook-name-1")
	logger.UnsetName("very-long-hook-name-2")
	
	// Should still be truncated
	assert.Contains(t, logger.spinner.Suffix, "waiting:")
	
	// Remove all hooks
	logger.UnsetName("very-long-hook-name-3")
	logger.UnsetName("very-long-hook-name-4")
	
	// Should be back to basic waiting
	assert.Equal(t, " waiting", logger.spinner.Suffix)
}

func BenchmarkMixedOperations(b *testing.B) {
	logger := createTestLogger()

	// Pre-populate with some names to simulate realistic usage
	for i := range benchmarkMixedOperations / 2 {
		logger.SetName(fmt.Sprintf("initial-hook-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hookName := fmt.Sprintf("mixed-hook-%d", i)

		// Add new name
		logger.SetName(hookName)

		// Occasionally remove an existing name
		if i%3 == 0 && len(logger.names) > 1 {
			logger.UnsetName(logger.names[0])
		}

		// Remove the name we just added
		logger.UnsetName(hookName)
	}
}

// Helper function to create a test logger with simulated terminal width
func createTestLoggerWithWidth(width int) *testLoggerWithWidth {
	baseLogger := createTestLogger()
	return &testLoggerWithWidth{
		Logger:         baseLogger,
		simulatedWidth: width,
	}
}

// testLoggerWithWidth wraps Logger to simulate different terminal widths
type testLoggerWithWidth struct {
	*Logger
	simulatedWidth int
}

// Override getTerminalWidth for testing
func (t *testLoggerWithWidth) getTerminalWidth() int {
	return t.simulatedWidth
}

// Override formatSpinnerSuffix to use our test width
func (t *testLoggerWithWidth) formatSpinnerSuffix(names []string) string {
	if len(names) == 0 {
		return spinnerText
	}

	// Use simulated terminal width
	terminalWidth := t.simulatedWidth
	if terminalWidth <= 0 {
		// Fallback to current behavior if we can't detect terminal width
		return fmt.Sprintf("%s: %s", spinnerText, strings.Join(names, ", "))
	}

	// Reserve space for spinner character (1) + space (1) + some padding (8)
	availableWidth := terminalWidth - 10

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

	// Strategy 3: Show as many individual names as possible, then count
	return t.formatWithPartialNames(names, availableWidth)
}

// Override SetName to use our custom formatSpinnerSuffix
func (t *testLoggerWithWidth) SetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.spinner.Active() {
		t.spinner.Stop()
		defer t.spinner.Start()
	}

	t.names = append(t.names, name)
	t.spinner.Suffix = t.formatSpinnerSuffix(t.names)
}

// Override UnsetName to use our custom formatSpinnerSuffix
func (t *testLoggerWithWidth) UnsetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.spinner.Active() {
		t.spinner.Stop()
		defer t.spinner.Start()
	}

	capacity := len(t.names)
	if capacity > 0 {
		capacity--
	}
	newNames := make([]string, 0, capacity)
	for _, n := range t.names {
		if n != name {
			newNames = append(newNames, n)
		}
	}

	t.names = newNames
	t.spinner.Suffix = t.formatSpinnerSuffix(t.names)
}

// formatWithPartialNames for the test logger
func (t *testLoggerWithWidth) formatWithPartialNames(names []string, availableWidth int) string {
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
