package config

import "github.com/evilmartians/lefthook/internal/git"

func isSkip(gitState git.State, value interface{}) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return typedValue == gitState.Step
	case []interface{}:
		for _, state := range typedValue {
			switch typedState := state.(type) {
			case string:
				if typedState == gitState.Step {
					return true
				}
			case map[string]interface{}:
				if typedState["ref"].(string) == gitState.Branch {
					return true
				}
			}
		}
	}
	return false
}
