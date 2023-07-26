package log

import (
	"fmt"
	"testing"
)

func TestSkipSetting(t *testing.T) {
	for i, tt := range [...]struct {
		settings []string
		results  map[string]bool
	}{
		{
			settings: []string{},
			results:  map[string]bool{},
		},
		{
			settings: []string{"failure", "execution"},
			results: map[string]bool{
				"failure":   true,
				"execution": true,
			},
		},
		{
			settings: []string{
				"meta",
				"summary",
				"success",
				"failure",
				"skips",
				"execution",
				"execution_out",
				"execution_info",
        "empty_summary"
			},
			results: map[string]bool{
				"meta":           true,
				"summary":        true,
				"success":        true,
				"failure":        true,
				"skips":          true,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
        "empty_summary": true,
			},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var settings SkipSettings

			for _, option := range tt.settings {
				(&settings).ApplySetting(option)
			}

			if settings.SkipMeta() != tt.results["meta"] {
				t.Errorf("expected SkipMeta to be %v", tt.results["meta"])
			}

			if settings.SkipSuccess() != tt.results["success"] {
				t.Errorf("expected SkipSuccess to be %v", tt.results["success"])
			}

			if settings.SkipFailure() != tt.results["failure"] {
				t.Errorf("expected SkipFailure to be %v", tt.results["failure"])
			}

			if settings.SkipSummary() != tt.results["summary"] {
				t.Errorf("expected SkipSummary to be %v", tt.results["summary"])
			}

			if settings.SkipExecution() != tt.results["execution"] {
				t.Errorf("expected SkipExecution to be %v", tt.results["execution"])
			}

			if settings.SkipExecutionOutput() != tt.results["execution_out"] {
				t.Errorf("expected SkipExecutionOutput to be %v", tt.results["execution_out"])
			}

			if settings.SkipExecutionInfo() != tt.results["execution_info"] {
				t.Errorf("expected SkipExecutionInfo to be %v", tt.results["execution_info"])
			}

      if settings.SkipEmptySummary() != tt.results["empty-summary"] {
        t.Errorf("expected SkipEmptySummary to be %v", tt.results["empty-summary"])
      }

			if settings.SkipSkips() != tt.results["skips"] {
				t.Errorf("expected SkipSkips to be %v", tt.results["skip"])
			}
		})
	}
}
