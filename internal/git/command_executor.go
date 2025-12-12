package git

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

// CommandExecutor provides some methods that take some effect on execution and/or result data.
type CommandExecutor struct {
	mu            *sync.Mutex
	cmd           system.Command
	root          string
	maxCmdLen     int
	onlyDebugLogs bool
	noTrimOut     bool

	// Do not print error messages
	silent bool
}

// NewExecutor returns an object that executes given commands in the OS.
func NewExecutor(cmd system.Command) *CommandExecutor {
	return &CommandExecutor{
		mu:        new(sync.Mutex),
		cmd:       cmd,
		maxCmdLen: system.MaxCmdLen(),
	}
}

func (c CommandExecutor) WithoutEnvs(envs ...string) CommandExecutor {
	c.cmd = c.cmd.WithoutEnvs(envs...)
	return c
}

func (c CommandExecutor) Silent() CommandExecutor {
	c.silent = true
	return c
}

func (c CommandExecutor) OnlyDebugLogs() CommandExecutor {
	c.onlyDebugLogs = true
	return c
}

func (c CommandExecutor) WithoutTrim() CommandExecutor {
	c.noTrimOut = true
	return c
}

// Cmd runs plain string command.
func (c CommandExecutor) Cmd(cmd []string) (string, error) {
	out, err := c.execute(cmd, c.root)
	if err != nil {
		return "", err
	}

	if !c.noTrimOut {
		out = strings.TrimSpace(out)
	}

	return out, nil
}

// BatchedCmd runs the command with any number of appended arguments batched in chunks to match the OS limits.
func (c CommandExecutor) BatchedCmd(cmd []string, args []string) (string, error) {
	result := strings.Builder{}

	argsBatched := batchByLength(args, c.maxCmdLen-len(cmd))
	for i, batch := range argsBatched {
		out, err := c.Cmd(append(cmd, batch...))
		if err != nil {
			return "", fmt.Errorf("error in batch %d: %w", i, err)
		}
		result.WriteString(strings.TrimRight(out, "\n"))
		result.WriteString("\n")
	}

	return result.String(), nil
}

// CmdLines runs plain string command, returns its output split by newline.
func (c CommandExecutor) CmdLines(cmd []string) ([]string, error) {
	out, err := c.Cmd(cmd)
	if err != nil {
		return nil, err
	}

	return strings.Split(out, "\n"), nil
}

// CmdLinesWithinFolder runs plain string command, returns its output split by newline.
func (c CommandExecutor) CmdLinesWithinFolder(cmd []string, folder string) ([]string, error) {
	root := filepath.Join(c.root, folder)
	out, err := c.execute(cmd, root)
	if err != nil {
		return nil, err
	}

	if !c.noTrimOut {
		out = strings.TrimSpace(out)
	}

	return strings.Split(out, "\n"), nil
}

func (c CommandExecutor) execute(cmd []string, root string) (string, error) {
	if len(cmd) > 0 && cmd[0] == "git" {
		// Preventing Git lock issues for all Git commands
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err := c.cmd.Run(cmd, root, system.NullReader, stdout, stderr)
	outString := stdout.String()
	errString := stderr.String()

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("git: ", strings.Join(cmd, " ")).
		Add("out: ", outString).
		Log()

	if err != nil {
		if len(errString) > 0 {
			logLevel := log.ErrorLevel
			if c.onlyDebugLogs || c.silent {
				logLevel = log.DebugLevel
			}
			log.Builder(logLevel, "> ").
				Add("", strings.Join(cmd, " ")).
				Add("", errString).
				Log()
		}
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
