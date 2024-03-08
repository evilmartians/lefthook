package log

import (
	"fmt"
	"testing"
)

func TestSkipSetting(t *testing.T) {
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
			results: map[string]bool{},
		},
		{
			tags:     "",
			settings: true,
			results: map[string]bool{
				"failure": true,
			},
		},
		{
			tags:     "meta,summary,success,skips,empty_summary",
			settings: nil,
			results: map[string]bool{
				"failure":        true,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
			},
		},
	} { //nolint:dupl // In next versions the `skip_settings_test` will be removed
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var settings SkipSettings

			(&settings).ApplySettings(tt.tags, tt.settings)

			if settings.LogMeta() != tt.results["meta"] {
				t.Errorf("expected LogMeta to be %v", tt.results["meta"])
			}

			if settings.LogSuccess() != tt.results["success"] {
				t.Errorf("expected LogSuccess to be %v", tt.results["success"])
			}

			if settings.LogFailure() != tt.results["failure"] {
				t.Errorf("expected LogFailure to be %v", tt.results["failure"])
			}

			if settings.LogSummary() != tt.results["summary"] {
				t.Errorf("expected LogSummary to be %v", tt.results["summary"])
			}

			if settings.LogExecution() != tt.results["execution"] {
				t.Errorf("expected LogExecution to be %v", tt.results["execution"])
			}

			if settings.LogExecutionOutput() != tt.results["execution_out"] {
				t.Errorf("expected LogExecutionOutput to be %v", tt.results["execution_out"])
			}

			if settings.LogExecutionInfo() != tt.results["execution_info"] {
				t.Errorf("expected LogExecutionInfo to be %v", tt.results["execution_info"])
			}

			if settings.LogEmptySummary() != tt.results["empty_summary"] {
				t.Errorf("expected LogEmptySummary to be %v", tt.results["empty_summary"])
			}

			if settings.LogSkips() != tt.results["skips"] {
				t.Errorf("expected LogSkips to be %v", tt.results["skip"])
			}
		})
	}
}
