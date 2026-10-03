package wrapper

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

type mockCmd struct{}

func (m mockCmd) WithoutEnvs(...string) system.Command { return mockCmd{} }
func (m mockCmd) Run(cmd []string, root string, in io.Reader, out io.Writer, errOut io.Writer) error {
	for _, str := range cmd {
		_, _ = out.Write([]byte(str))
		_, _ = out.Write([]byte("\n"))
	}

	return nil
}

func TestCmd_batchedCmd(t *testing.T) {
	assert := assert.New(t)

	c := NewCmd(mockCmd{}, loggertest.New())
	c.maxCmdLen = 2

	out, err := c.batchedCmd([]string{"hello"}, []string{"1", "2", "3", "4"})
	assert.NoError(err)

	assert.Equal("hello\n1\nhello\n2\nhello\n3\nhello\n4", out)
}
