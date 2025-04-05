package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/evilmartians/lefthook/internal/log"

	"github.com/mattn/go-isatty"
	"github.com/mattn/go-tty"
)

const plainSh = "sh"
const fullPathSh = `sh.exe`

type CommandExecutor struct{}
type executeArgs struct {
	in   io.Reader
	out  io.Writer
	envs []string
	root string
}

func (e CommandExecutor) Execute(ctx context.Context, opts Options, in io.Reader, out io.Writer) error {
	if opts.Interactive && !isatty.IsTerminal(os.Stdin.Fd()) {
		tty, err := tty.Open()
		if err == nil {
			defer tty.Close()
			in = tty.Input()
		} else {
			log.Errorf("Couldn't enable TTY input: %s\n", err)
		}
	}

	root, _ := filepath.Abs(opts.Root)
	envs := make([]string, len(opts.Env))
	for name, value := range opts.Env {
		envs = append(
			envs,
			fmt.Sprintf("%s=%s", name, os.ExpandEnv(value)),
		)
	}
	switch log.Colors() {
	case log.ColorOn:
		envs = append(envs, "CLICOLOR_FORCE=true")
	case log.ColorOff:
		envs = append(envs, "NO_COLOR=true")
	}

	args := &executeArgs{
		in:   in,
		out:  out,
		envs: envs,
		root: root,
	}

	for _, command := range opts.Commands {
		if err := e.execute(ctx, command, args); err != nil {
			return err
		}
	}

	return nil
}

func (e CommandExecutor) execute(ctx context.Context, cmdstr string, args *executeArgs) error {
	var sh string
	// Git hooks always setup GIT_INDEX env variable so here we check if we are in
	// a Git hook and can use `sh` without specifying the full path. This should cover most use cases.
	if len(os.Getenv("GIT_INDEX_FILE")) != 0 {
		sh = plainSh
	} else {
		// In case you call `lefthook run ...` from the terminal
		sh = fullPathSh
	}

	cmdStrQuoted := strings.ReplaceAll(cmdstr, "\"", "\\\"")
	cmdLine := "\"" + sh + "\"" + " -c " + "\"" + cmdStrQuoted + "\""
	log.Debug("[lefthook] run: ", cmdLine)

	command := exec.CommandContext(ctx, sh)
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: cmdLine,
	}
	command.Dir = args.root
	command.Env = append(os.Environ(), args.envs...)

	command.Stdout = args.out
	command.Stdin = args.in
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return err
	}

	defer func() { _ = command.Process.Kill() }()

	return command.Wait()
}
