package config

import "github.com/evilmartians/lefthook/internal/git"

func isSkip(gitSkipState git.State, value interface{}) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return git.State(typedValue) == gitSkipState
	case []interface{}:
		for _, gitState := range typedValue {
			if git.State(gitState.(string)) == gitSkipState {
				return true
			}
		}
	}
	return false
}
