package wrapper

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

// Cmd provides some methods that take some effect on execution and/or result data.
type Cmd struct {
	mu      *sync.Mutex
	logger  *logger.Logger
	command system.Command

	// Execute command in the specific directory
	root string

	// Split one command into multiple, respecting supported max command length
	maxCmdLen int

	// Print all logs in Debug level
	onlyDebugLogs bool

	// Do not trim leading and ending spaces
	noTrimOut bool
}

// NewCmd returns an object that executes given commands in the OS.
func NewCmd(command system.Command, logger *logger.Logger) *Cmd {
	return &Cmd{
		mu:        new(sync.Mutex),
		logger:    logger,
		command:   command,
		maxCmdLen: system.MaxCmdLen(),
	}
}

func (c Cmd) WithoutEnvs(envs ...string) Cmd {
	c.command = c.command.WithoutEnvs(envs...)
	return c
}

func (c Cmd) OnlyDebugLogs() Cmd {
	c.onlyDebugLogs = true
	return c
}

func (c Cmd) WithoutTrim() Cmd {
	c.noTrimOut = true
	return c
}

// cmd runs plain string command.
func (c Cmd) cmd(cmd []string) (string, error) {
	out, err := c.execute(cmd, c.root)
	if err != nil {
		return "", err
	}

	if !c.noTrimOut {
		out = strings.TrimSpace(out)
	}

	return out, nil
}

// batchedCmd runs the command with any number of appended arguments batched in chunks to match the OS limits.
func (c Cmd) batchedCmd(cmd []string, args []string) (string, error) {
	result := strings.Builder{}

	argsBatched := batchByLength(args, c.maxCmdLen-len(cmd))
	for i, batch := range argsBatched {
		out, err := c.cmd(append(cmd, batch...))
		if err != nil {
			return "", fmt.Errorf("error in batch %d: %w", i, err)
		}
		result.WriteString(strings.TrimRight(out, "\n"))
		result.WriteString("\n")
	}

	return strings.TrimRight(result.String(), "\n"), nil
}

// cmdLines runs plain string command, returns its output split by newline.
func (c Cmd) cmdLines(cmd []string) ([]string, error) {
	return c.cmdLinesRelative(cmd, "") // relative to current root
}

// cmdLinesWithinFolder runs plain string command, returns its output split by newline.
func (c Cmd) cmdLinesRelative(cmd []string, folder string) ([]string, error) {
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

func (c Cmd) execute(cmd []string, root string) (string, error) {
	if len(cmd) > 0 && cmd[0] == "git" {
		// Preventing Git lock issues for all Git commands
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err := c.command.Run(cmd, root, system.NullReader, stdout, stderr)
	outString := stdout.String()
	errString := stderr.String()

	logger.NewBuilder(c.logger).
		WithLevel(logger.LevelDebug).
		WithPrefix("[lefthook] ").
		WriteLines("git: ", strings.Join(cmd, " ")).
		WriteLines("out: ", outString).
		Log()

	if err != nil {
		if len(errString) > 0 {
			builder := logger.NewBuilder(c.logger).
				WithLevel(logger.LevelError).
				WithPrefix("> ")

			if c.onlyDebugLogs {
				builder = builder.WithLevel(logger.LevelDebug)
			}

			builder.
				WriteLines("", strings.Join(cmd, " ")).
				WriteLines("", errString).
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
