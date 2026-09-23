package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/logger"
)

const lefthookBinEnv = "LEFTHOOK_BIN"

var (
	executablePath = os.Executable
	evalSymlinks   = filepath.EvalSymlinks
)

// withLefthookEnv prepends the running lefthook binary's directory to PATH and
// sets LEFTHOOK_BIN so nested hook commands like `run: lefthook validate` work
// when git invokes hooks with a stripped PATH (see issue #1518).
func withLefthookEnv(env []string, log *logger.ExecutionLogger) []string {
	exe, err := executablePath()
	if err != nil {
		if log != nil {
			log.Warnf("Could not enrich hook PATH with lefthook binary: %v\n", err)
		}
		return env
	}
	if exe == "" {
		if log != nil {
			log.Warnf("Could not enrich hook PATH: empty executable path\n")
		}
		return env
	}
	exe, err = evalSymlinks(exe)
	if err != nil {
		if log != nil {
			log.Warnf("Could not resolve lefthook binary path: %v\n", err)
		}
		return env
	}
	return enrichEnvWithLefthookBinary(env, exe)
}

func enrichEnvWithLefthookBinary(env []string, exe string) []string {
	binDir := filepath.Dir(exe)

	hasBin := false
	hasPath := false
	out := make([]string, 0, len(env)+2)

	for _, entry := range env {
		if strings.HasPrefix(entry, lefthookBinEnv+"=") {
			hasBin = true
			out = append(out, entry)
			continue
		}
		if pathVal, ok := strings.CutPrefix(entry, "PATH="); ok {
			hasPath = true
			if !pathContainsDir(pathVal, binDir) {
				out = append(out, "PATH="+binDir+string(os.PathListSeparator)+pathVal)
			} else {
				out = append(out, entry)
			}
			continue
		}
		out = append(out, entry)
	}

	if !hasBin {
		out = append(out, fmt.Sprintf("%s=%s", lefthookBinEnv, exe))
	}
	if !hasPath {
		out = append(out, "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	return out
}

func pathContainsDir(pathEnv, dir string) bool {
	for _, part := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if part == dir {
			return true
		}
	}
	return false
}
