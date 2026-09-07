package command

// HookExitError is returned from `lefthook run` when hooks fail. Main uses
// Code as the process exit status so agent integrations (Claude Code Stop /
// PreToolUse) can block on exit code 2.
type HookExitError struct {
	Code int
}

func (e *HookExitError) Error() string {
	return ""
}
