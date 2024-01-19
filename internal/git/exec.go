package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
)

type Exec interface {
	SetRootPath(root string)
	Cmd(cmd []string) (string, error)
	CmdLines(cmd []string) ([]string, error)
}

type OsExec struct {
	root string
}

// NewOsExec returns an object that executes given commands
// in the OS.
func NewOsExec() *OsExec {
	return &OsExec{}
}

func (o *OsExec) SetRootPath(root string) {
	o.root = root
}

// Cmd runs plain string command. Trims spaces around output.
func (o *OsExec) Cmd(cmd []string) (string, error) {
	out, err := o.rawExecArgs(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (o *OsExec) CmdLines(cmd []string) ([]string, error) {
	out, err := o.rawExecArgs(cmd)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

// rawExecArgs executes git command with LEFTHOOK=0 in order
// to prevent calling subsequent lefthook hooks.
func (o *OsExec) rawExecArgs(args []string) (string, error) {
	log.Debug("[lefthook] cmd: ", args)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = o.root
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")

	out, err := cmd.CombinedOutput()
	log.Debug("[lefthook] dir: ", o.root)
	log.Debug("[lefthook] err: ", err)
	log.Debug("[lefthook] out: ", string(out))
	if err != nil {
		return "", err
	}

	return string(out), nil
}
