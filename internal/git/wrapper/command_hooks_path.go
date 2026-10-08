package wrapper

import "strings"

// CommandHooksPath returns core.hooksPath if and only if it is set at command
// scope (i.e. via `git -c core.hooksPath=...`). It returns the empty string
// for any file-based scope (local, global, system, worktree) or when the key
// is not configured at all.
//
// --show-scope requires git >= 2.26; lefthook's minimum is 2.31 so no version
// guard is necessary. On any error (key absent exits 1, unexpected output) we
// return "" to preserve base behavior.
func (w *Wrapper) CommandHooksPath() string {
	lines, err := w.cmd.cmdLines([]string{
		"git", "config", "--show-scope", "--get", "core.hooksPath",
	})
	if err != nil {
		return ""
	}
	// Output format: "<scope>\t<value>" for the single highest-precedence entry.
	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "command\t"); ok {
			return after
		}
	}
	return ""
}
