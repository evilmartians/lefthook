package git

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/system"
)

type gitCmd struct {
	cases map[string]string
}

func (g gitCmd) WithoutEnvs(...string) system.Command {
	return g
}

func (g gitCmd) Run(cmd []string, _root string, _in io.Reader, out io.Writer, _errOut io.Writer) error {
	res, ok := g.cases[(strings.Join(cmd, " "))]
	if !ok {
		return errors.New("doesn't exist")
	}

	_, err := out.Write([]byte(res))
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
					mu: new(sync.Mutex),
					cmd: gitCmd{
						cases: map[string]string{
							"git status --short --porcelain": tt.gitOut,
						},
					},
				},
			}
			repository.Setup()

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

func TestChangeset(t *testing.T) {
	for i, tt := range [...]struct {
		name, gitStatusOut, gitHashOut string
		pathsToHash                    []string
		result                         map[string]string
	}{
		{
			name:   "no changes",
			result: map[string]string{},
		},
		{
			name:         "modified file",
			gitStatusOut: " M modified.txt",
			gitHashOut:   "123456",
			pathsToHash:  []string{"modified.txt"},
			result: map[string]string{
				"modified.txt": "123456",
			},
		},
		{
			name:         "deleted file",
			gitStatusOut: "D  deleted.txt",
			result: map[string]string{
				"deleted.txt": "deleted",
			},
		},
		{
			name:         "new file",
			gitStatusOut: "?? new.txt",
			gitHashOut:   "654321",
			pathsToHash:  []string{"new.txt"},
			result: map[string]string{
				"new.txt": "654321",
			},
		},
		{
			name:         "new dir",
			gitStatusOut: "?? new-dir/",
			pathsToHash:  []string{},
			result: map[string]string{
				"new-dir/": "directory",
			},
		},
		{
			name: "mixed changes",
			gitStatusOut: `M  modified.txt
 D deleted.txt
?? new.txt
?? new-dir/
RM old-file -> new-file`,
			gitHashOut:  "123456\n654321\n758213",
			pathsToHash: []string{"modified.txt", "new.txt", "new-file"},
			result: map[string]string{
				"modified.txt": "123456",
				"deleted.txt":  "deleted",
				"new.txt":      "654321",
				"new-dir/":     "directory",
				"new-file":     "758213",
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			gitCmds := map[string]string{
				"git status --short --porcelain": tt.gitStatusOut,
			}

			if len(tt.pathsToHash) > 0 {
				cmd := append([]string{"git", "hash-object", "--"}, tt.pathsToHash...)
				gitCmds[strings.Join(cmd, " ")] = tt.gitHashOut
			}

			repository := &Repository{
				Git: &CommandExecutor{
					mu: new(sync.Mutex),
					cmd: gitCmd{
						cases: gitCmds,
					},
					maxCmdLen: 7000,
				},
			}
			repository.Setup()

			changeset, err := repository.Changeset()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if len(changeset) != len(tt.result) {
				t.Errorf("expected %d files, but %d returned", len(tt.result), len(changeset))
			}

			for file, hash := range tt.result {
				if changeset[file] != hash {
					t.Errorf("expected hash %s for file %s, but got %s", hash, file, changeset[file])
				}
			}
		})
	}
}
