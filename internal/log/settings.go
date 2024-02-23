package log

import (
	"strings"
)

const (
	meta = 1 << iota
	success
	failure
	summary
	skips
	execution
	executionOutput
	executionInfo
	emptySummary
	enableAll = ^0 // Set all bits as 1
)

type Settings int16

func NewSettings() SettingsInterface {
	var s Settings
	return &s
}

func (s *Settings) ApplySettings(tags string, output interface{}) {
	if tags == "" && (output == nil || output == "") {
		s.enableAll(true)
		return
	}

	if val, ok := output.(bool); ok {
		s.enableAll(val)
		return
	}

	if options, ok := output.([]interface{}); ok {
		for _, option := range options {
			if optStr, ok := option.(string); ok {
				s.applySetting(optStr)
			}
		}
	}

	if tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			s.applySetting(tag)
		}
	}
}

func (s *Settings) applySetting(setting string) {
	switch setting {
	case "meta":
		*s |= meta
	case "success":
		*s |= success
	case "failure":
		*s |= failure
	case "summary":
		*s |= summary
	case "skips":
		*s |= skips
	case "execution":
		*s |= execution
	case "execution_out":
		*s |= executionOutput
	case "execution_info":
		*s |= executionInfo
	case "empty_summary":
		*s |= emptySummary
	}
}

func (s *Settings) enableAll(val bool) {
	if val {
		*s = enableAll // Enable all params
	} else {
		*s |= skipFailure // Disable all params
	}
}

// Checks the state of params.
func (s Settings) isEnable(option int16) bool {
	return int16(s)&option != 0
}

// Using `SkipX` to maintain backward compatibility.
func (s Settings) SkipSuccess() bool {
	return !s.isEnable(success)
}

func (s Settings) SkipFailure() bool {
	return !s.isEnable(failure)
}

func (s Settings) SkipSummary() bool {
	return !s.isEnable(summary)
}

func (s Settings) SkipMeta() bool {
	return !s.isEnable(meta)
}

func (s Settings) SkipExecution() bool {
	return !s.isEnable(execution)
}

func (s Settings) SkipExecutionOutput() bool {
	return !s.isEnable(executionOutput)
}

func (s Settings) SkipExecutionInfo() bool {
	return !s.isEnable(executionInfo)
}

func (s Settings) SkipSkips() bool {
	return !s.isEnable(skips)
}

func (s Settings) SkipEmptySummary() bool {
	return !s.isEnable(emptySummary)
}
