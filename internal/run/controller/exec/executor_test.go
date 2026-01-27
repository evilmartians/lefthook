package exec

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	executor := CommandExecutor{}
	ctx := context.Background()

	t.Run("command executes successfully", func(t *testing.T) {
		opts := Options{
			Commands: []string{"echo hello"},
		}
		out := new(bytes.Buffer)

		err := executor.Execute(ctx, opts, strings.NewReader(""), out)
		assert.NoError(t, err)
	})

	t.Run("command with timeout field executes normally", func(t *testing.T) {
		opts := Options{
			Commands: []string{"echo hello"},
			Timeout:  "5s", // Timeout is passed but handled by controller, not executor
		}
		out := new(bytes.Buffer)

		err := executor.Execute(ctx, opts, strings.NewReader(""), out)
		assert.NoError(t, err)
	})
}
