package log

import (
	"fmt"
	"testing"
)

func TestSetting(t *testing.T) {
	for i, tt := range [...]struct {
		tags     string
		settings interface{}
		results  map[string]bool
	}{
		{
			tags:     "",
			settings: []interface{}{},
			results: map[string]bool{
				"meta":           true,
				"summary":        true,
				"success":        true,
				"failure":        true,
				"skips":          true,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
				"empty_summary":  true,
			},
		},
		{
			tags:     "",
			settings: false,
			results: map[string]bool{
				"meta":           true,
				"summary":        true,
				"success":        true,
				"failure":        false,
				"skips":          true,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
				"empty_summary":  true,
			},
		},
		{
			tags:     "",
			settings: []interface{}{"failure", "execution"},
			results: map[string]bool{
				"meta":           true,
				"summary":        true,
				"success":        true,
				"failure":        false,
				"skips":          true,
				"execution":      false,
				"execution_out":  true,
				"execution_info": true,
				"empty_summary":  true,
			},
		},
		{
			tags: "",
			settings: []interface{}{
				"meta",
				"summary",
				"success",
				"failure",
				"skips",
				"execution",
				"execution_out",
				"execution_info",
				"empty_summary",
			},
			results: map[string]bool{
				"meta":           false,
				"summary":        false,
				"success":        false,
				"failure":        false,
				"skips":          false,
				"execution":      false,
				"execution_out":  false,
				"execution_info": false,
				"empty_summary":  false,
			},
		},
		{
			tags:     "",
			settings: true,
			results: map[string]bool{
				"meta":           false,
				"summary":        false,
				"success":        false,
				"failure":        false,
				"skips":          false,
				"execution":      false,
				"execution_out":  false,
				"execution_info": false,
				"empty_summary":  false,
			},
		},
		{
			tags:     "meta,summary,success,skips,empty_summary",
			settings: nil,
			results: map[string]bool{
				"meta":           false,
				"summary":        false,
				"success":        false,
				"failure":        true,
				"skips":          false,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
				"empty_summary":  false,
			},
		},
	} { //nolint:dupl // In next versions the `skip_settings_test` will be removed
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var settings OutputSettings

			(&settings).ApplySettings(tt.tags, tt.settings)

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

			if settings.SkipEmptySummary() != tt.results["empty_summary"] {
				t.Errorf("expected SkipEmptySummary to be %v", tt.results["empty_summary"])
			}

			if settings.SkipSkips() != tt.results["skips"] {
				t.Errorf("expected SkipSkips to be %v", tt.results["skip"])
			}
		})
	}
}
