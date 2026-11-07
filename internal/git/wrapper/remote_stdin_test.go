package wrapper

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

type recordingCmd struct {
	cmds []string
	ins  []io.Reader
}

func (c *recordingCmd) WithoutEnvs(...string) system.Command { return c }

func (c *recordingCmd) Run(command []string, _ string, in io.Reader, _ io.Writer, _ io.Writer) error {
	c.cmds = append(c.cmds, strings.Join(command, " "))
	c.ins = append(c.ins, in)

	return nil
}

func Test_defaultRemoteStdin(t *testing.T) {
	origTerm := stdinIsTerminal
	origOpen := openConsole
	t.Cleanup(func() {
		stdinIsTerminal = origTerm
		openConsole = origOpen
	})

	t.Run("terminal-stdin", func(t *testing.T) {
		stdinIsTerminal = func() bool { return true }
		opened := false
		openConsole = func() (io.Reader, func(), error) {
			opened = true

			return nil, nil, errors.New("console should stay closed")
		}

		in, cleanup := defaultRemoteStdin()
		cleanup()

		if in != os.Stdin {
			t.Errorf("defaultRemoteStdin() = %v, want os.Stdin", in)
		}
		if opened {
			t.Error("defaultRemoteStdin() opened a console while stdin is a terminal")
		}
	})

	t.Run("hook-pipe", func(t *testing.T) {
		stdinIsTerminal = func() bool { return false }
		sentinel := &bytes.Buffer{}
		openConsole = func() (io.Reader, func(), error) {
			return sentinel, func() {}, nil
		}

		in, cleanup := defaultRemoteStdin()
		cleanup()

		if in != io.Reader(sentinel) {
			t.Errorf("defaultRemoteStdin() = %v, want console", in)
		}
		if in == os.Stdin {
			t.Error("defaultRemoteStdin() attached os.Stdin")
		}
	})

	t.Run("no-console", func(t *testing.T) {
		stdinIsTerminal = func() bool { return false }
		openConsole = func() (io.Reader, func(), error) {
			return nil, nil, errors.New("no tty")
		}

		in, cleanup := defaultRemoteStdin()
		cleanup()

		if in != system.NullReader {
			t.Errorf("defaultRemoteStdin() = %v, want NullReader", in)
		}
	})
}

func TestWrapper_remoteStdin(t *testing.T) {
	orig := remoteStdin
	t.Cleanup(func() { remoteStdin = orig })

	sentinel := &bytes.Buffer{}
	remoteStdin = func() (io.Reader, func()) {
		return sentinel, func() {}
	}

	fs := afero.NewMemMapFs()
	log := loggertest.New()

	t.Run("clone", func(t *testing.T) {
		rec := &recordingCmd{}
		w := New(fs, rec, log)

		err := w.Clone(CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"})
		if err != nil {
			t.Fatalf("Clone() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != io.Reader(sentinel) {
			t.Errorf("Clone() stdin = %v, want terminal", rec.ins)
		}
	})

	t.Run("fetch-not-checkout", func(t *testing.T) {
		rec := &recordingCmd{}
		w := New(fs, rec, log)

		err := w.Fetch("main", "/repo")
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if len(rec.ins) != 2 {
			t.Fatalf("Fetch() commands = %v, want fetch and checkout", rec.cmds)
		}
		if rec.ins[0] != io.Reader(sentinel) {
			t.Errorf("fetch stdin = %v, want terminal", rec.ins[0])
		}
		if rec.ins[1] != system.NullReader {
			t.Errorf("checkout stdin = %v, want NullReader", rec.ins[1])
		}
	})

	t.Run("pull", func(t *testing.T) {
		rec := &recordingCmd{}
		w := New(fs, rec, log)

		err := w.Pull("/repo")
		if err != nil {
			t.Fatalf("Pull() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != io.Reader(sentinel) {
			t.Errorf("Pull() stdin = %v, want terminal", rec.ins)
		}
	})

	t.Run("other-git-unchanged", func(t *testing.T) {
		rec := &recordingCmd{}
		w := New(fs, rec, log)

		_, err := w.Version()
		if err != nil {
			t.Fatalf("Version() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != system.NullReader {
			t.Errorf("Version() stdin = %v, want NullReader", rec.ins)
		}
	})
}
