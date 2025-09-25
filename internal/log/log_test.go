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

func TestLogger_SetName(t *testing.T) {
	tests := []struct {
		name           string
		initialNames   []string
		nameToAdd      string
		expectedNames  []string
		expectedSuffix string
	}{
		{
			name:           "add first name",
			initialNames:   []string{},
			nameToAdd:      "test-hook",
			expectedNames:  []string{"test-hook"},
			expectedSuffix: " waiting: test-hook",
		},
		{
			name:           "add second name",
			initialNames:   []string{"first-hook"},
			nameToAdd:      "second-hook",
			expectedNames:  []string{"first-hook", "second-hook"},
			expectedSuffix: " waiting: first-hook, second-hook",
		},
		{
			name:           "add multiple names",
			initialNames:   []string{"hook1", "hook2"},
			nameToAdd:      "hook3",
			expectedNames:  []string{"hook1", "hook2", "hook3"},
			expectedSuffix: " waiting: hook1, hook2, hook3",
		},
		{
			name:           "add empty name",
			initialNames:   []string{"existing"},
			nameToAdd:      "",
			expectedNames:  []string{"existing", ""},
			expectedSuffix: " waiting: existing, ",
		},
		{
			name:           "add duplicate name",
			initialNames:   []string{"hook1"},
			nameToAdd:      "hook1",
			expectedNames:  []string{"hook1", "hook1"},
			expectedSuffix: " waiting: hook1, hook1",
		},
		{
			name:           "add name with spaces",
			initialNames:   []string{},
			nameToAdd:      "hook with spaces",
			expectedNames:  []string{"hook with spaces"},
			expectedSuffix: " waiting: hook with spaces",
		},
		{
			name:           "add name with unicode",
			initialNames:   []string{},
			nameToAdd:      "🥊-hook",
			expectedNames:  []string{"🥊-hook"},
			expectedSuffix: " waiting: 🥊-hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := createTestLogger()
			
			// Set up initial state
			logger.names = make([]string, len(tt.initialNames))
			copy(logger.names, tt.initialNames)
			
			// Call SetName
			logger.SetName(tt.nameToAdd)
			
			// Verify names slice
			assert.Equal(t, tt.expectedNames, logger.names)
			
			// Verify spinner suffix
			assert.Equal(t, tt.expectedSuffix, logger.spinner.Suffix)
		})
	}
}

func TestLogger_UnsetName(t *testing.T) {
	tests := []struct {
		name           string
		initialNames   []string
		nameToRemove   string
		expectedNames  []string
		expectedSuffix string
	}{
		{
			name:           "remove only name",
			initialNames:   []string{"test-hook"},
			nameToRemove:   "test-hook",
			expectedNames:  []string{},
			expectedSuffix: " waiting",
		},
		{
			name:           "remove first of two names",
			initialNames:   []string{"first-hook", "second-hook"},
			nameToRemove:   "first-hook",
			expectedNames:  []string{"second-hook"},
			expectedSuffix: " waiting: second-hook",
		},
		{
			name:           "remove second of two names",
			initialNames:   []string{"first-hook", "second-hook"},
			nameToRemove:   "second-hook",
			expectedNames:  []string{"first-hook"},
			expectedSuffix: " waiting: first-hook",
		},
		{
			name:           "remove middle name",
			initialNames:   []string{"hook1", "hook2", "hook3"},
			nameToRemove:   "hook2",
			expectedNames:  []string{"hook1", "hook3"},
			expectedSuffix: " waiting: hook1, hook3",
		},
		{
			name:           "remove non-existent name",
			initialNames:   []string{"hook1", "hook2"},
			nameToRemove:   "hook3",
			expectedNames:  []string{"hook1", "hook2"},
			expectedSuffix: " waiting: hook1, hook2",
		},
		{
			name:           "remove from empty list",
			initialNames:   []string{},
			nameToRemove:   "hook1",
			expectedNames:  []string{},
			expectedSuffix: " waiting",
		},
		{
			name:           "remove empty name",
			initialNames:   []string{"hook1", "", "hook2"},
			nameToRemove:   "",
			expectedNames:  []string{"hook1", "hook2"},
			expectedSuffix: " waiting: hook1, hook2",
		},
		{
			name:           "remove all duplicates",
			initialNames:   []string{"hook1", "hook1", "hook2"},
			nameToRemove:   "hook1",
			expectedNames:  []string{"hook2"},
			expectedSuffix: " waiting: hook2",
		},
		{
			name:           "remove unicode name",
			initialNames:   []string{"🥊-hook", "normal-hook"},
			nameToRemove:   "🥊-hook",
			expectedNames:  []string{"normal-hook"},
			expectedSuffix: " waiting: normal-hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := createTestLogger()
			
			// Set up initial state
			logger.names = make([]string, len(tt.initialNames))
			copy(logger.names, tt.initialNames)
			
			// Call UnsetName
			logger.UnsetName(tt.nameToRemove)
			
			// Verify names slice
			assert.Equal(t, tt.expectedNames, logger.names)
			
			// Verify spinner suffix
			assert.Equal(t, tt.expectedSuffix, logger.spinner.Suffix)
		})
	}
}

func TestLogger_SetName_UnsetName_Integration(t *testing.T) {
	logger := createTestLogger()
	
	// Start with empty state
	assert.Equal(t, []string{}, logger.names)
	assert.Equal(t, " waiting", logger.spinner.Suffix)
	
	// Add first name
	logger.SetName("hook1")
	assert.Equal(t, []string{"hook1"}, logger.names)
	assert.Equal(t, " waiting: hook1", logger.spinner.Suffix)
	
	// Add second name
	logger.SetName("hook2")
	assert.Equal(t, []string{"hook1", "hook2"}, logger.names)
	assert.Equal(t, " waiting: hook1, hook2", logger.spinner.Suffix)
	
	// Add third name
	logger.SetName("hook3")
	assert.Equal(t, []string{"hook1", "hook2", "hook3"}, logger.names)
	assert.Equal(t, " waiting: hook1, hook2, hook3", logger.spinner.Suffix)
	
	// Remove middle name
	logger.UnsetName("hook2")
	assert.Equal(t, []string{"hook1", "hook3"}, logger.names)
	assert.Equal(t, " waiting: hook1, hook3", logger.spinner.Suffix)
	
	// Remove first name
	logger.UnsetName("hook1")
	assert.Equal(t, []string{"hook3"}, logger.names)
	assert.Equal(t, " waiting: hook3", logger.spinner.Suffix)
	
	// Remove last name
	logger.UnsetName("hook3")
	assert.Equal(t, []string{}, logger.names)
	assert.Equal(t, " waiting", logger.spinner.Suffix)
}

func TestLogger_LongHookNames(t *testing.T) {
	logger := createTestLogger()
	
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
	assert.Equal(t, longNames, logger.names)
	
	// Verify the suffix contains all names (this is the current behavior that causes the issue)
	expectedSuffix := " waiting: " + strings.Join(longNames, ", ")
	assert.Equal(t, expectedSuffix, logger.spinner.Suffix)
	
	// Document the current problematic behavior
	t.Logf("Current suffix length: %d characters", len(logger.spinner.Suffix))
	t.Logf("This would cause wrapping issues in terminals narrower than %d columns", len(logger.spinner.Suffix))
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	logger := createTestLogger()
	
	const numGoroutines = 10
	const numOperations = 50
	
	var wg sync.WaitGroup
	
	// Start goroutines that concurrently add and remove names
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
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
	assert.Equal(t, []string{}, logger.names)
	assert.Equal(t, " waiting", logger.spinner.Suffix)
}

func TestLogger_SpinnerActiveHandling(t *testing.T) {
	logger := createTestLogger()
	
	// Test that SetName and UnsetName don't panic when spinner is active
	logger.spinner.Start()
	initialActive := logger.spinner.Active()
	
	// SetName should handle active spinner without panicking
	logger.SetName("test-hook")
	assert.Equal(t, []string{"test-hook"}, logger.names)
	assert.Equal(t, " waiting: test-hook", logger.spinner.Suffix)
	
	// UnsetName should handle active spinner without panicking
	logger.UnsetName("test-hook")
	assert.Equal(t, []string{}, logger.names)
	assert.Equal(t, " waiting", logger.spinner.Suffix)
	
	// Clean up
	logger.spinner.Stop()
	
	// Document the behavior for future reference
	t.Logf("Spinner was initially active: %v", initialActive)
}

func TestGlobalSetNameUnsetName(t *testing.T) {
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
	assert.Equal(t, []string{"global-hook"}, std.names)
	assert.Equal(t, " waiting: global-hook", std.spinner.Suffix)
	
	// Test global UnsetName
	UnsetName("global-hook")
	assert.Equal(t, []string{}, std.names)
	assert.Equal(t, " waiting", std.spinner.Suffix)
}

// Helper function to create a test logger
func createTestLogger() *Logger {
	return &Logger{
		level:   InfoLevel,
		out:     &bytes.Buffer{},
		colors:  ColorOff,
		names:   []string{},
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

// Benchmark tests to understand performance characteristics
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
	for i := 0; i < 1000; i++ {
		logger.names = append(logger.names, fmt.Sprintf("hook-%d", i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.UnsetName(fmt.Sprintf("hook-%d", i%1000))
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
