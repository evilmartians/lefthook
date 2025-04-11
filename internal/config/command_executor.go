package config

import (
	"bytes"
	"runtime"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/system"
)

// commandExecutor implements execution of a skip checks passed in a `run` option.
type commandExecutor struct {
	cmd system.Command
}

// cmd runs plain string command in a subshell returning the success of it.
func (c *commandExecutor) execute(commandLine string) bool {
	if commandLine == "" {
		return false
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"powershell", "-Command", commandLine}
	} else {
		args = []string{"sh", "-c", commandLine}
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err := c.cmd.Run(args, "", system.NullReader, stdout, stderr)

	b := log.Builder(log.DebugLevel, "[lefthook] ").
		Add("run: ", strings.Join(args, " ")).
		Add("out: ", stdout.String()).
		Add("err: ", stderr.String())

	if err != nil {
		b.Add("!:   ", err.Error())
	}

	b.Log()

	return err == nil
}
