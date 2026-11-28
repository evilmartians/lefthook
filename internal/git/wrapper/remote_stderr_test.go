package wrapper

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

const (
	promptWait       = 2 * time.Second
	passphrasePrompt = "Enter passphrase for key:\n"
)

// promptHoldCmd writes a passphrase prompt to stderr and then blocks, so a
// test can observe the prompt while git is still waiting for it.
type promptHoldCmd struct {
	release chan struct{}
	started chan struct{}
	in      io.Reader
	prompt  string
}

func (c *promptHoldCmd) WithoutEnvs(...string) system.Command { return c }

func (c *promptHoldCmd) Run(_ []string, _ string, in io.Reader, _ io.Writer, errOut io.Writer) error {
	c.in = in
	if _, err := io.WriteString(errOut, c.prompt); err != nil {
		return err
	}
	close(c.started)
	<-c.release

	return nil
}

type captureCmd struct {
	cmds    []string
	ins     []io.Reader
	errOuts []io.Writer
}

func (c *captureCmd) WithoutEnvs(...string) system.Command { return c }

func (c *captureCmd) Run(command []string, _ string, in io.Reader, _ io.Writer, errOut io.Writer) error {
	c.cmds = append(c.cmds, strings.Join(command, " "))
	c.ins = append(c.ins, in)
	c.errOuts = append(c.errOuts, errOut)

	return nil
}

// terminalStderr points os.Stderr at a pipe and forces the terminal check.
// The returned reader sees whatever git writes to the terminal.
func terminalStderr(t *testing.T, terminal bool) (*os.File, *os.File) {
	t.Helper()

	origTerm := stderrIsTerminal
	origErr := os.Stderr
	stderrIsTerminal = func() bool { return terminal }

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = pw
	t.Cleanup(func() {
		stderrIsTerminal = origTerm
		os.Stderr = origErr
		_ = pw.Close()
		_ = pr.Close()
	})

	return pr, pw
}

func readClosed(t *testing.T, pr *os.File, pw *os.File) string {
	t.Helper()

	// Drop the write end so ReadAll can finish.
	_ = pw.Close()
	got, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	return string(got)
}

func TestClone_promptVisibleBeforeGitExits(t *testing.T) {
	pr, _ := terminalStderr(t, true)

	release := make(chan struct{})
	started := make(chan struct{})
	cmd := &promptHoldCmd{release: release, started: started, prompt: passphrasePrompt}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})

	done := make(chan error, 1)
	go func() {
		w := New(afero.NewMemMapFs(), cmd, loggertest.New())
		done <- w.Clone(CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"})
	}()

	select {
	case <-started:
	case <-time.After(promptWait):
		t.Fatal("clone did not start")
	}

	buf := make([]byte, len(passphrasePrompt))
	_ = pr.SetReadDeadline(time.Now().Add(promptWait))
	n, readErr := pr.Read(buf)
	if !bytes.Contains(buf[:n], []byte("Enter passphrase for key:")) {
		t.Fatalf("prompt not visible before git exits: %q err=%v", buf[:n], readErr)
	}

	select {
	case err := <-done:
		t.Fatalf("Clone returned before the prompt was answered: %v", err)
	default:
	}

	if cmd.in != system.NullReader {
		t.Fatalf("stdin = %v, want discarded", cmd.in)
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
}

func TestWrapper_remoteStderr(t *testing.T) {
	const prompt = "Enter passphrase for key: "

	t.Run("clone-terminal", func(t *testing.T) {
		pr, pw := terminalStderr(t, true)
		rec := &captureCmd{}
		w := New(afero.NewMemMapFs(), rec, loggertest.New())

		err := w.Clone(CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"})
		if err != nil {
			t.Fatalf("Clone() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != system.NullReader {
			t.Fatalf("stdin = %v, want discarded", rec.ins)
		}
		if _, err = io.WriteString(rec.errOuts[0], prompt); err != nil {
			t.Fatalf("write prompt: %v", err)
		}
		if got := readClosed(t, pr, pw); !strings.Contains(got, prompt) {
			t.Fatalf("terminal = %q, want prompt", got)
		}
	})

	t.Run("clone-not-terminal", func(t *testing.T) {
		pr, pw := terminalStderr(t, false)
		rec := &captureCmd{}
		w := New(afero.NewMemMapFs(), rec, loggertest.New())

		err := w.Clone(CloneArgs{Root: "/root", Dest: "dest", Url: "https://example.com/repo.git"})
		if err != nil {
			t.Fatalf("Clone() error = %v", err)
		}
		if _, err = io.WriteString(rec.errOuts[0], prompt); err != nil {
			t.Fatalf("write prompt: %v", err)
		}
		if got := readClosed(t, pr, pw); got != "" {
			t.Fatalf("terminal = %q, want empty when stderr is not a terminal", got)
		}
	})

	t.Run("fetch-not-checkout", func(t *testing.T) {
		pr, pw := terminalStderr(t, true)
		rec := &captureCmd{}
		w := New(afero.NewMemMapFs(), rec, loggertest.New())

		err := w.Fetch("main", "/repo")
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if len(rec.errOuts) != 2 {
			t.Fatalf("Fetch() commands = %v, want fetch and checkout", rec.cmds)
		}
		if rec.ins[0] != system.NullReader || rec.ins[1] != system.NullReader {
			t.Fatalf("stdin = %v, want both discarded", rec.ins)
		}
		if _, err = io.WriteString(rec.errOuts[0], "fetch-prompt"); err != nil {
			t.Fatalf("write fetch prompt: %v", err)
		}
		if _, err = io.WriteString(rec.errOuts[1], "checkout-hidden"); err != nil {
			t.Fatalf("write checkout stderr: %v", err)
		}
		got := readClosed(t, pr, pw)
		if !strings.Contains(got, "fetch-prompt") {
			t.Fatalf("terminal = %q, want fetch prompt", got)
		}
		if strings.Contains(got, "checkout-hidden") {
			t.Fatalf("terminal = %q, checkout stderr must stay buffered", got)
		}
	})

	t.Run("pull", func(t *testing.T) {
		pr, pw := terminalStderr(t, true)
		rec := &captureCmd{}
		w := New(afero.NewMemMapFs(), rec, loggertest.New())

		err := w.Pull("/repo")
		if err != nil {
			t.Fatalf("Pull() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != system.NullReader {
			t.Fatalf("stdin = %v, want discarded", rec.ins)
		}
		if _, err = io.WriteString(rec.errOuts[0], prompt); err != nil {
			t.Fatalf("write prompt: %v", err)
		}
		if got := readClosed(t, pr, pw); !strings.Contains(got, prompt) {
			t.Fatalf("terminal = %q, want prompt", got)
		}
	})

	t.Run("other-git-unchanged", func(t *testing.T) {
		pr, pw := terminalStderr(t, true)
		rec := &captureCmd{}
		w := New(afero.NewMemMapFs(), rec, loggertest.New())

		_, err := w.Version()
		if err != nil {
			t.Fatalf("Version() error = %v", err)
		}
		if len(rec.ins) != 1 || rec.ins[0] != system.NullReader {
			t.Fatalf("stdin = %v, want discarded", rec.ins)
		}
		if _, err = io.WriteString(rec.errOuts[0], "version-hidden"); err != nil {
			t.Fatalf("write stderr: %v", err)
		}
		if got := readClosed(t, pr, pw); got != "" {
			t.Fatalf("terminal = %q, other git commands must keep stderr buffered", got)
		}
	})
}
