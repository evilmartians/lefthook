package logger

import "fmt"

type ExecutionSettings int16
type ExecutionSetting = ExecutionSettings

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

func (s *ExecutionSettings) apply(setting ExecutionSetting) {
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
		setting = LogSummary
	case "empty_summary":
		setting = LogEmptySummary
	case "skips":
		setting = LogSkips
	case "execution":
		setting = LogExecution
	case "execution_output":
		setting = LogExecutionOutput
	case "execution_info":
		setting = LogExecutionInfo
	case "setup":
		setting = LogSetup
	default:
		return 0, fmt.Errorf("Unknown output setting: %v", name)
	}
	return setting, nil
}
