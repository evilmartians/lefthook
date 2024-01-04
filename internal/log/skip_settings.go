package log

import "strings"

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

type SkipSettings int16

func (s *SkipSettings) ApplySetting(setting string) {
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

func (s *SkipSettings) ApplySkipSettings(tags string, skipOutput interface{}) {
	switch typedSkipOutput := skipOutput.(type) {
	case bool:
		s.SkipAll(typedSkipOutput)
	case []string:
		if tags != "" {
			typedSkipOutput = append(typedSkipOutput, strings.Split(tags, ",")...)
		}
		for _, skipOption := range typedSkipOutput {
			s.ApplySetting(skipOption)
		}
	}
}

func (s *SkipSettings) SkipAll(val bool) {
	if val {
		*s = skipAll &^ skipFailure
	} else {
		*s = 0
	}
}

func (s SkipSettings) SkipSuccess() bool {
	return s.doSkip(skipSuccess)
}

func (s SkipSettings) SkipFailure() bool {
	return s.doSkip(skipFailure)
}

func (s SkipSettings) SkipSummary() bool {
	return s.doSkip(skipSummary)
}

func (s SkipSettings) SkipMeta() bool {
	return s.doSkip(skipMeta)
}

func (s SkipSettings) SkipExecution() bool {
	return s.doSkip(skipExecution)
}

func (s SkipSettings) SkipExecutionOutput() bool {
	return s.doSkip(skipExecutionOutput)
}

func (s SkipSettings) SkipExecutionInfo() bool {
	return s.doSkip(skipExecutionInfo)
}

func (s SkipSettings) SkipSkips() bool {
	return s.doSkip(skipSkips)
}

func (s SkipSettings) SkipEmptySummary() bool {
	return s.doSkip(skipEmptySummary)
}

func (s SkipSettings) doSkip(option int16) bool {
	return int16(s)&option != 0
}
