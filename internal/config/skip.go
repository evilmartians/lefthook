package config

import (
	"github.com/gobwas/glob"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
)

type SkipChecker struct {
	Executor Exec
}

func NewSkipChecker(executor Exec) *SkipChecker {
	if executor == nil {
		executor = NewOsExec()
	}

	return &SkipChecker{Executor: executor}
}

func (sc *SkipChecker) DoSkip(gitState git.State, skip interface{}, only interface{}) bool {
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

func (sc *SkipChecker) matches(gitState git.State, value interface{}) bool {
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

func (sc *SkipChecker) matchesSlices(gitState git.State, slice []interface{}) bool {
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

func (sc *SkipChecker) matchesRef(gitState git.State, typedState map[string]interface{}) bool {
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

func (sc *SkipChecker) matchesCommands(typedState map[string]interface{}) bool {
	commandLine, ok := typedState["run"].(string)
	if !ok {
		return false
	}

	log.Debug("[lefthook] skip/only cmd: ", commandLine)

	return sc.Executor.Cmd(commandLine)
}
