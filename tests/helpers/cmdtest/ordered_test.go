package cmdtest

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

func TestOrderedCmd(t *testing.T) {
	var _ system.Command = (*OrderedCmd)(nil)

	errFailed := errors.New("failed")

	cmd := NewOrdered(
		t,
		[]Out{
			{Command: "A 1"},
			{Command: "B 2"},
			{Command: "C 3"},
			{Command: "D 4", Err: errFailed},
		},
	)
	_ = cmd.WithoutEnvs("OK")

	assert.NoError(t, cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.Run([]string{"B", "2"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.Run([]string{"C", "3"}, "", system.NullReader, io.Discard, io.Discard))
	assert.ErrorIs(t, cmd.Run([]string{"D", "4"}, "", system.NullReader, io.Discard, io.Discard), errFailed)
}
