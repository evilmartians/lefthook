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
)

const (
	disableAll = 0
	skipMeta   = (1 << iota)
	skipSuccess
	skipFailure
	skipSummary
	skipSkips
	skipExecution
	skipExecutionOutput
	skipExecutionInfo
	skipEmptySummary
)

type Settings interface {
	Apply(enableTags, disableTags string, enable, disable interface{})
	LogSuccess() bool
	LogFailure() bool
	LogSummary() bool
	LogMeta() bool
	LogExecution() bool
	LogExecutionOutput() bool
	LogExecutionInfo() bool
	LogSkips() bool
	LogEmptySummary() bool
}

type LogSettings struct {
	bitmap int16
}

func NewSettings() Settings {
	s := LogSettings{^disableAll}
	return &s
}

func (s *LogSettings) Apply(enableTags, disableTags string, enable, disable interface{}) {
	if enableTags == "" && disableTags == "" && (enable == nil || enable == "") && (disable == nil || disable == "") {
		s.enableAll()
		return
	}

	if enableOutput, ok := enable.(bool); ok && enableTags == "" && disableTags == "" {
		if enableOutput {
			s.enableAll()
		} else {
			s.disableAll()
		}
		return
	}

	if disableOutput, ok := disable.(bool); ok && enableTags == "" && disableTags == "" {
		if disableOutput {
			s.disableAll()
			return
		}
	}

	if enableOptions, ok := enable.([]interface{}); ok {
		if len(enableOptions) != 0 {
			s.bitmap = disableAll
		}

		for _, option := range enableOptions {
			if value, ok := option.(string); ok {
				s.enable(value)
			}
		}
	}

	if disableOptions, ok := disable.([]interface{}); ok {
		for _, option := range disableOptions {
			if value, ok := option.(string); ok {
				s.disable(value)
			}
		}
	}

	if enableTags != "" {
		s.bitmap = disableAll

		for _, tag := range strings.Split(enableTags, ",") {
			s.enable(tag)
		}
	}

	if disableTags != "" {
		for _, tag := range strings.Split(disableTags, ",") {
			s.disable(tag)
		}
	}
}

func (s *LogSettings) enable(setting string) {
	switch setting {
	case "meta":
		s.bitmap |= meta
	case "success":
		s.bitmap |= success
	case "failure":
		s.bitmap |= failure
	case "summary":
		s.bitmap |= summary | success | failure
	case "skips":
		s.bitmap |= skips
	case "execution":
		s.bitmap |= execution | executionOutput | executionInfo
	case "execution_out":
		s.bitmap |= executionOutput | execution
	case "execution_info":
		s.bitmap |= executionInfo | execution
	case "empty_summary":
		s.bitmap |= emptySummary
	}
}

func (s *LogSettings) disable(setting string) {
	switch setting {
	case "meta":
		s.bitmap &= ^meta
	case "success":
		s.bitmap &= ^success
	case "failure":
		s.bitmap &= ^failure
	case "summary":
		s.bitmap &= ^summary & ^success & ^failure
	case "skips":
		s.bitmap &= ^skips
	case "execution":
		s.bitmap &= ^execution & ^executionOutput & ^executionInfo
	case "execution_out":
		s.bitmap &= ^executionOutput
	case "execution_info":
		s.bitmap &= ^executionInfo
	case "empty_summary":
		s.bitmap &= ^emptySummary
	}
}

func (s *LogSettings) enableAll() {
	s.bitmap = ^disableAll
}

func (s *LogSettings) disableAll() {
	s.bitmap = failure
}

// Checks the state of params.
func (s LogSettings) isEnable(option int16) bool {
	return s.bitmap&option != 0
}

func (s LogSettings) LogSuccess() bool {
	return s.isEnable(success)
}

func (s LogSettings) LogFailure() bool {
	return s.isEnable(failure)
}

func (s LogSettings) LogSummary() bool {
	return s.isEnable(summary)
}

func (s LogSettings) LogMeta() bool {
	return s.isEnable(meta)
}

func (s LogSettings) LogExecution() bool {
	return s.isEnable(execution)
}

func (s LogSettings) LogExecutionOutput() bool {
	return s.isEnable(executionOutput)
}

func (s LogSettings) LogExecutionInfo() bool {
	return s.isEnable(executionInfo)
}

func (s LogSettings) LogSkips() bool {
	return s.isEnable(skips)
}

func (s LogSettings) LogEmptySummary() bool {
	return s.isEnable(emptySummary)
}
