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
func (sc *skipChecker) Check(state func() git.State, skip interface{}, only interface{}) bool {
	if skip == nil && only == nil {
		return false
	}

	if skip != nil {
		if sc.matches(state, skip) {
			return true
		}
	}

	if only != nil {
		return !sc.matches(state, only)
	}

	return false
}

func (sc *skipChecker) matches(state func() git.State, value interface{}) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return typedValue == state().State
	case []interface{}:
		return sc.matchesSlices(state, typedValue)
	}
	return false
}

func (sc *skipChecker) matchesSlices(gitState func() git.State, slice []interface{}) bool {
	for _, state := range slice {
		switch typedState := state.(type) {
		case string:
			if typedState == gitState().State {
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

func (sc *skipChecker) matchesRef(state func() git.State, typedState map[string]interface{}) bool {
	ref, ok := typedState["ref"].(string)
	if !ok {
		return false
	}

	branch := state().Branch
	if ref == branch {
		return true
	}

	g := glob.MustCompile(ref)

	return g.Match(branch)
}

func (sc *skipChecker) matchesCommands(typedState map[string]interface{}) bool {
	commandLine, ok := typedState["run"].(string)
	if !ok {
		return false
	}

	result := sc.exec.execute(commandLine)

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("skip/only: ", commandLine).
		Add("result:    ", result).
		Log()

	return result
}
