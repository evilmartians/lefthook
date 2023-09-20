package git

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type GitMock struct {
	cases map[string]string
}

func (g GitMock) SetRootPath(_root string) {}

func (g GitMock) Cmd(cmd string) (string, error) {
	res, err := g.RawCmd(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res), nil
}

func (g GitMock) CmdArgs(args ...string) (string, error) {
	return g.Cmd(strings.Join(args, " "))
}

func (g GitMock) CmdLines(cmd string) ([]string, error) {
	res, err := g.Cmd(cmd)
	if err != nil {
		return nil, err
	}

	return strings.Split(res, "\n"), nil
}

func (g GitMock) RawCmd(cmd string) (string, error) {
	res, ok := g.cases[cmd]
	if !ok {
		return "", errors.New("doesn't exist")
	}

	return res, nil
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
				Git: GitMock{
					cases: map[string]string{
						"git status --short": tt.gitOut,
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
