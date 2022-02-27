package lefthook

import (
	"fmt"
)

const (
	outputMeta    string = "meta"
	outputSummary string = "summary"
	outputSuccess string = "success"
)

type outputDisablerFn func(out ...interface{}) string

func outputDisabler(output string, disabledOutputs []string, envDisabledOutputs []string) outputDisablerFn {
	if isOutputDisabled(output, disabledOutputs, envDisabledOutputs) {
		return func(...interface{}) string {
			return ""
		}
	}
	return func(out ...interface{}) string {
		return fmt.Sprint(out...)
	}
}

func isOutputDisabled(output string, disabledOutputs []string, envDisabledOutputs []string) bool {
	for _, disabledOutput := range disabledOutputs {
		if output == disabledOutput {
			return true
		}
	}
	for _, disabledOutput := range envDisabledOutputs {
		if output == disabledOutput {
			return true
		}
	}
	return false
}
