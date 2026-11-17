package wrapper

import (
	"bytes"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// stderrIsTerminal reports whether process stderr is a terminal.
// Tests replace it.
var stderrIsTerminal = func() bool {
	fd := os.Stderr.Fd()

	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// remoteStderr is the writer for git clone, fetch, and pull of config remotes.
//
// OpenSSH reads an SSH key passphrase from /dev/tty, not from the git process
// stdin (sshconnect2 load_identity_file calls read_passphrase with no
// RP_ALLOW_STDIN, which requires a tty). Discarding that stdin returns EOF at
// once, so it is not what makes lefthook install wait on "fetching remotes".
// The status spinner redraws over the tty prompt, and stderr was buffered until
// git exited, which it does not do while it is waiting. Copy stderr to the
// terminal as git writes it. Stdin stays discarded: lefthook run syncs the same
// remotes and must not consume the hook's stdin, and opening a console with
// go-tty would clear ECHO, ICANON and ICRNL for the whole fetch.
func remoteStderr(buf *bytes.Buffer) io.Writer {
	if !stderrIsTerminal() {
		return buf
	}

	return io.MultiWriter(os.Stderr, buf)
}
