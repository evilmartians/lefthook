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
const fullPathGitDirDefault = `C:\Program Files\Git`

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
		var err error

		sh, err = getFullPathSh()
		if err != nil {
			log.Errorf("Couldn't find sh.exe: %s\n", err)
			return err
		}
	}

	cmdStrQuoted := strings.ReplaceAll(strings.ReplaceAll(cmdstr, "\\", "\\\\"), "\"", "\\\"")
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

func getFullPathSh() (string, error) {
	var fullPathSh string
	gitbashDir, err := findExecutableDir("git-bash.exe")
	if err == nil {
		fullPathSh = filepath.Join(gitbashDir, "sh.exe")
		if _, err := os.Stat(fullPathSh); err == nil {
			return fullPathSh, nil
		}
		fullPathSh = filepath.Join(gitbashDir, "bin", "sh.exe")
		if _, err := os.Stat(fullPathSh); err == nil {
			return fullPathSh, nil
		}
	}

	gitDir, err := findExecutableDir("git.exe")
	if err == nil {
		baseDir := filepath.Dir(gitDir)
		fullPathSh = filepath.Join(baseDir, "bin", "sh.exe")
		if _, err := os.Stat(fullPathSh); err == nil {
			return fullPathSh, nil
		}
	}
	fullPathSh = filepath.Join(fullPathGitDirDefault, "bin", "sh.exe")
	if _, err := os.Stat(fullPathSh); err == nil {
		return fullPathSh, nil
	}
	return "", fmt.Errorf("sh.exe not found in PATH")
}
func findExecutableDir(cmdStr string) (string, error) {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	for _, dir := range paths {
		findPath := filepath.Join(dir, cmdStr)
		if _, err := os.Stat(findPath); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("%s not found in PATH", cmdStr)
}
