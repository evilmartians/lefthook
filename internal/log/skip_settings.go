package log

import (
	"strings"
)

const (
	skipMeta = 1 << iota
	skipSuccess
	skipFailure
	skipSummary
	skipSkips
	skipExecution
	skipExecutionOutput
	skipExecutionInfo
	skipEmptySummary
	skipAll = (1 << iota) - 1
)

// Deprecated: Use Settings instead.
type SkipSettings int16

// Deprecated: Use NewSettings instead.
func NewSkipSettings() Settings {
	var s SkipSettings
	return &s
}

func (s *SkipSettings) ApplySettings(tags string, skipOutput interface{}) {
	switch typedSkipOutput := skipOutput.(type) {
	case bool:
		s.skipAll(typedSkipOutput)
	case []interface{}:
		for _, skipOption := range typedSkipOutput {
			s.applySetting(skipOption.(string))
		}
	}

	if tags != "" {
		for _, skipOption := range strings.Split(tags, ",") {
			s.applySetting(skipOption)
		}
	}
}

func (s *SkipSettings) applySetting(setting string) {
	switch setting {
	case "meta":
		*s |= skipMeta
	case "success":
		*s |= skipSuccess
	case "failure":
		*s |= skipFailure
	case "summary":
		*s |= skipSummary
	case "skips":
		*s |= skipSkips
	case "execution":
		*s |= skipExecution
	case "execution_out":
		*s |= skipExecutionOutput
	case "execution_info":
		*s |= skipExecutionInfo
	case "empty_summary":
		*s |= skipEmptySummary
	}
}

func (s *SkipSettings) skipAll(val bool) {
	if val {
		*s = skipAll &^ skipFailure
	} else {
		*s = 0
	}
}

func (s SkipSettings) LogSuccess() bool {
	return !s.doSkip(skipSuccess)
}

func (s SkipSettings) LogFailure() bool {
	return !s.doSkip(skipFailure)
}

func (s SkipSettings) LogSummary() bool {
	return !s.doSkip(skipSummary)
}

func (s SkipSettings) LogMeta() bool {
	return !s.doSkip(skipMeta)
}

func (s SkipSettings) LogExecution() bool {
	return !s.doSkip(skipExecution)
}

func (s SkipSettings) LogExecutionOutput() bool {
	return !s.doSkip(skipExecutionOutput)
}

func (s SkipSettings) LogExecutionInfo() bool {
	return !s.doSkip(skipExecutionInfo)
}

func (s SkipSettings) LogSkips() bool {
	return !s.doSkip(skipSkips)
}

func (s SkipSettings) LogEmptySummary() bool {
	return !s.doSkip(skipEmptySummary)
}

func (s SkipSettings) doSkip(option int16) bool {
	return int16(s)&option != 0
}
