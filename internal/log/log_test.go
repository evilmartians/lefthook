package log

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/briandowns/spinner"
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
	for i := range b.N {
		logger.SetName(fmt.Sprintf("hook-%d", i))
	}
}

func BenchmarkUnsetName(b *testing.B) {
	logger := createTestLogger()

	// Pre-populate with names
	for i := range benchmarkPrePopulationSize {
		logger.names = append(logger.names, fmt.Sprintf("hook-%d", i))
	}

	b.ResetTimer()
	for i := range b.N {
		logger.UnsetName(fmt.Sprintf("hook-%d", i%benchmarkPrePopulationSize))
	}
}

func BenchmarkSetNameUnsetName(b *testing.B) {
	logger := createTestLogger()

	b.ResetTimer()
	for i := range b.N {
		hookName := fmt.Sprintf("hook-%d", i)
		logger.SetName(hookName)
		logger.UnsetName(hookName)
	}
}

func BenchmarkMixedOperations(b *testing.B) {
	logger := createTestLogger()

	// Pre-populate with some names to simulate realistic usage
	for i := range benchmarkMixedOperations / 2 {
		logger.SetName(fmt.Sprintf("initial-hook-%d", i))
	}

	b.ResetTimer()
	for i := range b.N {
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
