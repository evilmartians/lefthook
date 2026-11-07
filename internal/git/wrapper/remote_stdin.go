package wrapper

import (
	"bytes"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/mattn/go-tty"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

// remoteStdin is the reader passed to git clone, fetch, and pull while syncing
// config remotes. Tests replace it.
var remoteStdin = defaultRemoteStdin

// stdinIsTerminal reports whether process stdin is a terminal.
var stdinIsTerminal = func() bool {
	fd := os.Stdin.Fd()

	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// openConsole opens a terminal that is not the process stdin.
var openConsole = func() (io.Reader, func(), error) {
	term, err := tty.Open()
	if err != nil {
		return nil, nil, err
	}

	return term.Input(), func() { _ = term.Close() }, nil
}

// defaultRemoteStdin returns a terminal so an SSH key passphrase can be typed.
//
// lefthook run syncs the same remotes, and a hook may already own os.Stdin
// (pre-push refs, use_stdin jobs). That stream is attached only when it is
// already a terminal. Otherwise a separate console is opened, and if there is
// none, stdin stays discarded.
func defaultRemoteStdin() (io.Reader, func()) {
	if stdinIsTerminal() {
		return os.Stdin, func() {}
	}

	in, cleanup, err := openConsole()
	if err != nil {
		return system.NullReader, func() {}
	}

	return in, cleanup
}

// stderrForRemote copies git's stderr to the terminal when the attached
// reader is a terminal, and always keeps a buffer for the error log.
func stderrForRemote(in io.Reader, buf *bytes.Buffer) io.Writer {
	if !isTerminalFile(in) {
		return buf
	}

	return io.MultiWriter(os.Stderr, buf)
}

func isTerminalFile(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}

	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
