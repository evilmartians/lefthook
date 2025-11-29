package log

import (
	"strconv"
	"testing"
)

func TestSetting(t *testing.T) {
	for i, tt := range [...]struct {
		enableTags     string
		enableSettings any
		results        map[string]bool
	}{
		{
			enableTags:     "",
			enableSettings: []any{},
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
			enableTags:     "",
			enableSettings: false,
			results: map[string]bool{
				"failure": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []any{"success"},
			results: map[string]bool{
				"success": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []any{"summary"},
			results: map[string]bool{
				"summary": true,
				"success": true,
				"failure": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []any{"failure", "execution"},
			results: map[string]bool{
				"failure":        true,
				"execution":      true,
				"execution_info": true,
				"execution_out":  true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []any{"failure", "execution_out"},
			results: map[string]bool{
				"failure":       true,
				"execution":     true,
				"execution_out": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []any{"failure", "execution_info"},
			results: map[string]bool{
				"failure":        true,
				"execution":      true,
				"execution_info": true,
			},
		},
		{
			enableTags: "",
			enableSettings: []any{
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
			enableTags:     "",
			enableSettings: true,
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
			enableTags:     "meta,summary,skips,empty_summary",
			enableSettings: nil,
			results: map[string]bool{
				"meta":          true,
				"summary":       true,
				"success":       true,
				"failure":       true,
				"skips":         true,
				"empty_summary": true,
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			settings := NewSettings()
			settings.Apply(tt.enableTags, tt.enableSettings)

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
