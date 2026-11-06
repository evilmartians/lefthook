//go:build !windows

package command

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestNotifyHookContext_CancelsOnTerminationSignals(t *testing.T) {
	for name, sig := range map[string]syscall.Signal{
		"SIGINT":  syscall.SIGINT,
		"SIGTERM": syscall.SIGTERM,
		"SIGHUP":  syscall.SIGHUP,
	} {
		t.Run(name, func(t *testing.T) {
			ctx, stop := notifyHookContext(context.Background())
			defer stop()

			if err := syscall.Kill(syscall.Getpid(), sig); err != nil {
				t.Fatal(err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
				t.Fatalf("context was not cancelled by %s", name)
			}
		})
	}
}
