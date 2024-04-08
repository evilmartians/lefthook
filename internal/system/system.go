// system package contains OS-specific implementations.
package system

import (
	"os"
	"os/exec"

	"github.com/evilmartians/lefthook/internal/log"
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
