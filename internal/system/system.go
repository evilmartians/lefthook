// system package contains OS-specific implementations.
package system

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/evilmartians/lefthook/internal/log"
)

const (
	// https://serverfault.com/questions/69430/what-is-the-maximum-length-of-a-command-line-in-mac-os-x
	// https://support.microsoft.com/en-us/help/830473/command-prompt-cmd-exe-command-line-string-limitation
	// https://unix.stackexchange.com/a/120652
	maxCommandLengthDarwin  = 260000 // 262144
	maxCommandLengthWindows = 7000   // 8191, but see issues#655
	maxCommandLengthLinux   = 130000 // 131072
)

type Executor struct{}

// Execute executes git command with LEFTHOOK=0 in order
// to prevent calling subsequent lefthook hooks.
func (e Executor) Execute(args []string, root string) (string, error) {
	log.Debug("[lefthook] cmd: ", args)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")
	if len(root) > 0 {
		cmd.Dir = root
	}

	out, err := cmd.CombinedOutput()
	log.Debug("[lefthook] dir: ", root)
	log.Debug("[lefthook] err: ", err)
	log.Debug("[lefthook] out: ", string(out))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func MaxCmdLen() int {
	switch runtime.GOOS {
	case "windows":
		return maxCommandLengthWindows
	case "darwin":
		return maxCommandLengthDarwin
	default:
		return maxCommandLengthLinux
	}
}
