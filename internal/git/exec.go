package git

import (
	"os"
	"os/exec"
	"strings"
)

type Exec interface {
	Cmd(cmd string) (string, error)
	CmdArgs(args ...string) (string, error)
	CmdLines(cmd string) ([]string, error)
	RawCmd(cmd string) (string, error)
}

type OsExec struct{}

// NewOsExec returns an object that executes given commands
// in the OS.
func NewOsExec() *OsExec {
	return &OsExec{}
}

func (o *OsExec) Cmd(cmd string) (string, error) {
	args := strings.Split(cmd, " ")
	return o.CmdArgs(args...)
}

func (o *OsExec) CmdLines(cmd string) ([]string, error) {
	out, err := o.RawCmd(cmd)
	if err != nil {
		return nil, err
	}

	return strings.Split(out, "\n"), nil
}

func (o *OsExec) CmdArgs(args ...string) (string, error) {
	out, err := o.rawExecArgs(args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func (o *OsExec) RawCmd(cmd string) (string, error) {
	args := strings.Split(cmd, " ")
	return o.rawExecArgs(args...)
}

// rawExecArgs executes git command with LEFTHOOK=0 in order
// to prevent calling subsequent lefthook hooks.
func (o *OsExec) rawExecArgs(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
