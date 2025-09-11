package git

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/system"
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

func TestBatchedCmd(t *testing.T) {
	assert := assert.New(t)
	c := CommandExecutor{cmd: mockCmd{}, maxCmdLen: 2}
	out, err := c.BatchedCmd([]string{"hello"}, []string{"1", "2", "3", "4"})
	assert.NoError(err)

	assert.Equal("hello\n1\nhello\n2\nhello\n3\nhello\n4\n", out)
}
