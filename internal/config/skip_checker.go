package config

import (
	"github.com/gobwas/glob"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
)

type skipChecker struct {
	exec *commandExecutor
}

func NewSkipChecker(cmd system.Command) *skipChecker {
	return &skipChecker{&commandExecutor{cmd}}
}

// check returns the result of applying a skip/only setting which can be a branch, git state, shell command, etc.
func (sc *skipChecker) check(gitState git.State, skip interface{}, only interface{}) bool {
	if skip != nil {
		if sc.matches(gitState, skip) {
			return true
		}
	}

	if only != nil {
		return !sc.matches(gitState, only)
	}

	return false
}

func (sc *skipChecker) matches(gitState git.State, value interface{}) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return typedValue == gitState.Step
	case []interface{}:
		return sc.matchesSlices(gitState, typedValue)
	}
	return false
}

func (sc *skipChecker) matchesSlices(gitState git.State, slice []interface{}) bool {
	for _, state := range slice {
		switch typedState := state.(type) {
		case string:
			if typedState == gitState.Step {
				return true
			}
		case map[string]interface{}:
			if sc.matchesRef(gitState, typedState) {
				return true
			}

			if sc.matchesCommands(typedState) {
				return true
			}
		}
	}

	return false
}

func (sc *skipChecker) matchesRef(gitState git.State, typedState map[string]interface{}) bool {
	ref, ok := typedState["ref"].(string)
	if !ok {
		return false
	}

	if ref == gitState.Branch {
		return true
	}

	g := glob.MustCompile(ref)

	return g.Match(gitState.Branch)
}

func (sc *skipChecker) matchesCommands(typedState map[string]interface{}) bool {
	commandLine, ok := typedState["run"].(string)
	if !ok {
		return false
	}

	result := sc.exec.execute(commandLine)

	log.Debugf("[lefthook] skip/only cmd: %s, result: %t", commandLine, result)

	return result
}
