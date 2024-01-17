package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
)

type Exec interface {
	SetRootPath(root string)
	Cmd(cmd []string) (string, error)
	CmdLines(cmd []string) ([]string, error)
	CmdLinesWithinFolder(cmd []string, folder string) ([]string, error)
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
	out, err := o.rawExecArgs(cmd, "")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (o *OsExec) CmdLines(cmd []string) ([]string, error) {
	out, err := o.rawExecArgs(cmd, "")
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (o *OsExec) CmdLinesWithinFolder(cmd []string, folder string) ([]string, error) {
	out, err := o.rawExecArgs(cmd, folder)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

// rawExecArgs executes git command with LEFTHOOK=0 in order
// to prevent calling subsequent lefthook hooks.
func (o *OsExec) rawExecArgs(args []string, folder string) (string, error) {
	log.Debug("[lefthook] cmd: ", args)

	root := filepath.Join(o.root, folder)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")

	out, err := cmd.CombinedOutput()
	log.Debug("[lefthook] dir: ", root)
	log.Debug("[lefthook] err: ", err)
	log.Debug("[lefthook] out: ", string(out))
	if err != nil {
		return "", err
	}

	return string(out), nil
}
