package config

import (
	"bytes"
	"strings"

	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

// commandExecutor implements execution of a skip checks passed in a `run` option.
type commandExecutor struct {
	logger *logger.Logger
	cmd    system.Command
}

// cmd runs plain string command in a subshell returning the success of it.
func (c *commandExecutor) execute(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	sh, err := system.Sh()
	if err != nil {
		c.logger.Errorf("`sh` executable not found: %s\n", err)
		return false
	}

	args := []string{sh, "-c", commandLine}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err = c.cmd.Run(args, "", system.NullReader, stdout, stderr)

	b := logger.NewBuilder(c.logger).
		WithPrefix("[lefthook] ").
		WithLevel(logger.LevelDebug).
		WriteLines("run: ", strings.Join(args, " ")).
		WriteLines("out: ", stdout.String()).
		WriteLines("err: ", stderr.String())

	if err != nil {
		b.WriteLines("!:   ", err.Error())
	}

	b.Log()

	return err == nil
}
