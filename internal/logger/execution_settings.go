package logger

import "fmt"

type (
	ExecutionSettings int16
	ExecutionSetting  = ExecutionSettings
)

const (
	LogMeta            ExecutionSettings = 1 << iota // meta
	LogSuccess                                       // success
	LogFailure                                       // failure
	LogSummary                                       // summary
	LogEmptySummary                                  // empty_summary
	LogSkips                                         // skips
	LogExecution                                     // execution
	LogExecutionOutput                               // execution_output
	LogExecutionInfo                                 // execution_info
	LogSetup                                         // setup

	executionFull ExecutionSettings = ^0
	executionNone ExecutionSettings = 0
)

func NewExecutionSettings() *ExecutionSettings {
	return new(ExecutionSettings)
}

func (s *ExecutionSettings) enable(setting ExecutionSetting) {
	(*s) |= setting
}

func (s *ExecutionSettings) enabled(setting ExecutionSetting) bool {
	return ((*s) & setting) != 0
}

func nameToSetting(name string) (ExecutionSetting, error) {
	var setting ExecutionSetting
	switch name {
	case "meta":
		setting = LogMeta
	case "success":
		setting = LogSuccess
	case "failure":
		setting = LogFailure
	case "summary":
		setting = LogSummary | LogSuccess | LogFailure
	case "empty_summary":
		setting = LogEmptySummary
	case "skips":
		setting = LogSkips
	case "execution", "jobs":
		setting = LogExecution | LogExecutionOutput | LogExecutionInfo
	case "execution_out", "jobs_out":
		setting = LogExecution | LogExecutionOutput
	case "execution_info", "jobs_info":
		setting = LogExecution | LogExecutionInfo
	case "setup":
		setting = LogSetup
	default:
		return 0, fmt.Errorf("Unknown output setting: %#v", name)
	}
	return setting, nil
}
