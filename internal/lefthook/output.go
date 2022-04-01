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

func outputDisabler(output string, disabledOutputs []string) outputDisablerFn {
	if isOutputDisabled(output, disabledOutputs) {
		return func(...interface{}) string {
			return ""
		}
	}
	return func(out ...interface{}) string {
		return fmt.Sprint(out...)
	}
}

func isOutputDisabled(output string, disabledOutputs []string) bool {
	for _, disabledOutput := range disabledOutputs {
		if output == disabledOutput {
			return true
		}
	}
	return false
}
