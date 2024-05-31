package git

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

type gitCmd struct {
	cases map[string]string
}

func (g gitCmd) Run(cmd []string, _root string, _in io.Reader, out io.Writer) error {
	res, ok := g.cases[(strings.Join(cmd, " "))]
	if !ok {
		return errors.New("doesn't exist")
	}

	_, err := out.Write([]byte(strings.TrimSpace(res)))
	if err != nil {
		return err
	}

	return nil
}

func TestPartiallyStagedFiles(t *testing.T) {
	for i, tt := range [...]struct {
		name, gitOut string
		error        bool
		result       []string
	}{
		{
			gitOut: `RM old-file -> new file
M  staged
MM staged but changed
`,
			result: []string{"new file", "staged but changed"},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			repository := &Repository{
				Git: &CommandExecutor{
					cmd: gitCmd{
						cases: map[string]string{
							"git status --short --porcelain": tt.gitOut,
						},
					},
				},
			}

			files, err := repository.PartiallyStagedFiles()
			if tt.error && err != nil {
				t.Errorf("expected an error")
			}

			if len(files) != len(tt.result) {
				t.Errorf("expected %d files, but %d returned", len(tt.result), len(files))
			}

			for j, file := range files {
				if tt.result[j] != file {
					t.Errorf("file at index %d don't match: %s - %s", j, tt.result[j], file)
				}
			}
		})
	}
}
