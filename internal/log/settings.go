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

type Settings interface {
	ApplySettings(tags string, skipOutput interface{})
	SkipSuccess() bool
	SkipFailure() bool
	SkipSummary() bool
	SkipMeta() bool
	SkipExecution() bool
	SkipExecutionOutput() bool
	SkipExecutionInfo() bool
	SkipSkips() bool
	SkipEmptySummary() bool
}

type OutputSettings int16

func NewSettings() Settings {
	var s OutputSettings
	return &s
}

func (s *OutputSettings) ApplySettings(tags string, output interface{}) {
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

func (s *OutputSettings) applySetting(setting string) {
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

func (s *OutputSettings) enableAll(val bool) {
	if val {
		*s = enableAll // Enable all params
	} else {
		*s |= skipFailure // Disable all params
	}
}

// Checks the state of params.
func (s OutputSettings) isEnable(option int16) bool {
	return int16(s)&option != 0
}

// Using `SkipX` to maintain backward compatibility.
func (s OutputSettings) SkipSuccess() bool {
	return !s.isEnable(success)
}

func (s OutputSettings) SkipFailure() bool {
	return !s.isEnable(failure)
}

func (s OutputSettings) SkipSummary() bool {
	return !s.isEnable(summary)
}

func (s OutputSettings) SkipMeta() bool {
	return !s.isEnable(meta)
}

func (s OutputSettings) SkipExecution() bool {
	return !s.isEnable(execution)
}

func (s OutputSettings) SkipExecutionOutput() bool {
	return !s.isEnable(executionOutput)
}

func (s OutputSettings) SkipExecutionInfo() bool {
	return !s.isEnable(executionInfo)
}

func (s OutputSettings) SkipSkips() bool {
	return !s.isEnable(skips)
}

func (s OutputSettings) SkipEmptySummary() bool {
	return !s.isEnable(emptySummary)
}
