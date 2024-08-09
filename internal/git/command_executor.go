package git

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
)

// CommandExecutor provides some methods that take some effect on execution and/or result data.
type CommandExecutor struct {
	cmd  system.Command
	root string
}

// NewExecutor returns an object that executes given commands in the OS.
func NewExecutor(cmd system.Command) *CommandExecutor {
	return &CommandExecutor{cmd: cmd}
}

// Cmd runs plain string command. Trims spaces around output.
func (c CommandExecutor) Cmd(cmd []string) (string, error) {
	out, err := c.execute(cmd, c.root)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// BatchedCmd runs the command with any number of appended arguments batched in chunks to match the OS limits.
func (c CommandExecutor) BatchedCmd(cmd []string, args []string) (string, error) {
	maxlen := system.MaxCmdLen()
	result := strings.Builder{}

	argsBatched := batchByLength(args, maxlen-len(cmd))
	for i, batch := range argsBatched {
		out, err := c.Cmd(append(cmd, batch...))
		if err != nil {
			return "", fmt.Errorf("error in batch %d: %w", i, err)
		}
		result.WriteString(out)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (c CommandExecutor) CmdLines(cmd []string) ([]string, error) {
	out, err := c.execute(cmd, c.root)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (c CommandExecutor) CmdLinesWithinFolder(cmd []string, folder string) ([]string, error) {
	root := filepath.Join(c.root, folder)
	out, err := c.execute(cmd, root)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

func (c CommandExecutor) execute(cmd []string, root string) (string, error) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	err := c.cmd.Run(cmd, root, system.NullReader, out, errOut)
	outString := out.String()

	log.Debug("[lefthook] stdout: ", outString)
	errString := errOut.String()
	if len(errString) > 0 {
		log.Debug("[lefthook] stderr: ", errString)
	}

	return outString, err
}

func batchByLength(s []string, length int) [][]string {
	batches := make([][]string, 0)

	var acc, prev int
	for i := range s {
		acc += len(s[i])
		if acc > length {
			if i == prev {
				batches = append(batches, s[prev:i+1])
				prev = i + 1
			} else {
				batches = append(batches, s[prev:i])
				prev = i
			}
			acc = len(s[i])
		}
	}
	if acc > 0 {
		batches = append(batches, s[prev:])
	}

	return batches
}
