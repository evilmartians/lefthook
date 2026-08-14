package cmdtest

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

func TestTrackingCmd(t *testing.T) {
	var _ system.Command = (*TrackingCmd)(nil)
	var _ system.CommandWithContext = (*TrackingCmd)(nil)

	commands := make([]string, 0, 3)
	cb := func(command string, root string, _ io.Writer) error {
		commands = append(commands, command)
		return nil
	}
	cmd := NewTracking(cb)

	assert.NoError(t, cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.Run([]string{"B", "2"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.RunWithContext(t.Context(), []string{"C", "3"}, "", system.NullReader, io.Discard, io.Discard))

	assert.Equal(t, []string{"A 1", "B 2", "C 3"}, cmd.Commands)
	assert.Equal(t, []string{"A 1", "B 2", "C 3"}, commands)

	_ = cmd.WithoutEnvs("OK")
}
