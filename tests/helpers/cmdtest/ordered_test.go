package cmdtest

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

func TestOrderedCmd(t *testing.T) {
	var _ system.Command = (*OrderedCmd)(nil)

	cmd := NewOrdered(
		t,
		[]Out{
			{"A 1", ""},
			{"B 2", ""},
			{"C 3", ""},
		},
	)
	_ = cmd.WithoutEnvs("OK")

	assert.NoError(t, cmd.Run([]string{"A", "1"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.Run([]string{"B", "2"}, "", system.NullReader, io.Discard, io.Discard))
	assert.NoError(t, cmd.Run([]string{"C", "3"}, "", system.NullReader, io.Discard, io.Discard))
}
