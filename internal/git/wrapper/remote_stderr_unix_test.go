//go:build !windows

package wrapper

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestClone_passphrasePromptVisibleBeforeGitExits(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = master.Close()
		_ = slave.Close()
	})

	origStderr := os.Stderr
	os.Stderr = slave
	t.Cleanup(func() { os.Stderr = origStderr })

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
	got := make(chan string, 1)
	go func() {
		n, _ := master.Read(buf)
		got <- string(buf[:n])
	}()
	select {
	case text := <-got:
		if !bytes.Contains([]byte(text), []byte("Enter passphrase for key:")) {
			t.Fatalf("prompt not visible before git exits: %q", text)
		}
	case <-time.After(promptWait):
		t.Fatal("prompt not visible before git exits")
	}

	select {
	case err = <-done:
		t.Fatalf("Clone returned before the prompt was answered: %v", err)
	default:
	}

	if cmd.in != system.NullReader {
		t.Fatalf("stdin = %v, want discarded", cmd.in)
	}

	close(release)
	// Cleanup also closes release. Receiving here makes that close a no-op.
	if err = <-done; err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
}
