package log

import (
	"strconv"
	"testing"
)

func TestSetting(t *testing.T) {
	for i, tt := range [...]struct {
		enableTags, disableTags         string
		enableSettings, disableSettings interface{}
		results                         map[string]bool
	}{
		{
			enableTags:     "",
			enableSettings: []interface{}{},
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
			enableSettings: []interface{}{"success"},
			results: map[string]bool{
				"success": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []interface{}{"summary"},
			results: map[string]bool{
				"summary": true,
				"success": true,
				"failure": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []interface{}{"failure", "execution"},
			results: map[string]bool{
				"failure":        true,
				"execution":      true,
				"execution_info": true,
				"execution_out":  true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []interface{}{"failure", "execution_out"},
			results: map[string]bool{
				"failure":       true,
				"execution":     true,
				"execution_out": true,
			},
		},
		{
			enableTags:     "",
			enableSettings: []interface{}{"failure", "execution_info"},
			results: map[string]bool{
				"failure":        true,
				"execution":      true,
				"execution_info": true,
			},
		},
		{
			enableTags: "",
			enableSettings: []interface{}{
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
		{
			disableTags:     "",
			disableSettings: []interface{}{},
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
			disableTags:     "",
			disableSettings: false,
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
			disableTags:     "",
			disableSettings: []interface{}{"failure", "execution"},
			results: map[string]bool{
				"meta":           true,
				"summary":        true,
				"success":        true,
				"failure":        false,
				"skips":          true,
				"execution":      false,
				"execution_out":  false,
				"execution_info": false,
				"empty_summary":  true,
			},
		},
		{
			disableTags: "",
			disableSettings: []interface{}{
				"meta",
				"summary",
				"skips",
				"execution",
				"execution_out",
				"execution_info",
				"empty_summary",
			},
			results: map[string]bool{},
		},
		{
			disableTags:     "",
			disableSettings: true,
			results: map[string]bool{
				"failure": true,
			},
		},
		{
			disableTags:     "meta,summary,success,skips,empty_summary",
			disableSettings: nil,
			results: map[string]bool{
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
			},
		},
		{
			disableTags:     "meta,success,skips,empty_summary",
			disableSettings: nil,
			results: map[string]bool{
				"summary":        true,
				"failure":        true,
				"execution":      true,
				"execution_out":  true,
				"execution_info": true,
			},
		},
		{
			enableSettings:  true, // this takes precedence
			disableSettings: true,
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
			enableSettings:  []interface{}{"meta"},
			disableSettings: true, // this takes precedence
			results: map[string]bool{
				"failure": true,
			},
		},
		{
			enableSettings:  true,
			disableSettings: []interface{}{"meta"},
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
			enableSettings:  []interface{}{"summary", "execution"},
			disableSettings: []interface{}{"failure", "execution_out"},
			results: map[string]bool{
				"summary":        true,
				"success":        true,
				"execution":      true,
				"execution_info": true,
			},
		},
		{
			enableTags:      "summary,execution", // takes precedence
			disableSettings: []interface{}{"failure", "execution_out"},
			results: map[string]bool{
				"summary":        true,
				"success":        true,
				"failure":        true,
				"execution":      true,
				"execution_info": true,
				"execution_out":  true,
			},
		},
		{
			disableTags:    "summary,execution",
			enableSettings: []interface{}{"meta", "summary", "execution_info"},
			results: map[string]bool{
				"meta": true,
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			settings := NewSettings()
			settings.Apply(tt.enableTags, tt.disableTags, tt.enableSettings, tt.disableSettings)

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
