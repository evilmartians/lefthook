package git

import (
	"fmt"
	"sync"
	"testing"

	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
)

func TestPartiallyStagedFiles(t *testing.T) {
	for i, tt := range [...]struct {
		name   string
		git    []cmdtest.Out
		error  bool
		result []string
	}{
		{
			git: []cmdtest.Out{
				{
					Command: "git status --short --porcelain -z",
					Output: "RM new file\x00old-file\x00" +
						"M  staged\x00" +
						"MM staged but changed\x00",
				},
			},
			result: []string{"new file", "staged but changed"},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			repository := &Repository{
				Git: &CommandExecutor{
					mu:  new(sync.Mutex),
					cmd: cmdtest.NewOrdered(t, tt.git),
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
		name        string
		git         []cmdtest.Out
		pathsToHash []string
		result      map[string]string
	}{
		{
			name: "no changes",
			git: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: ""},
			},
			result: map[string]string{},
		},
		{
			name: "modified file",
			git: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: " M modified.txt\x00"},
				{Command: "git hash-object -- modified.txt", Output: "123456"},
			},
			pathsToHash: []string{"modified.txt"},
			result: map[string]string{
				"modified.txt": "123456",
			},
		},
		{
			name: "deleted file",
			git: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "D  deleted.txt\x00"},
			},
			result: map[string]string{
				"deleted.txt": "deleted",
			},
		},
		{
			name: "new file",
			git: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "?? new.txt\x00"},
				{Command: "git hash-object -- new.txt", Output: "654321"},
			},
			pathsToHash: []string{"new.txt"},
			result: map[string]string{
				"new.txt": "654321",
			},
		},
		{
			name: "new dir",
			git: []cmdtest.Out{
				{Command: "git status --short --porcelain -z", Output: "?? new-dir/\x00"},
			},
			pathsToHash: []string{},
			result: map[string]string{
				"new-dir/": "directory",
			},
		},
		{
			name: "mixed changes",
			git: []cmdtest.Out{
				{
					Command: "git status --short --porcelain -z",
					Output: "M  modified.txt\x00" +
						"CT copied to\x00copied from\x00" +
						" D deleted.txt\x00" +
						"?? new.txt\x00" +
						"?? new-dir/\x00" +
						"RM new-file\x00old-file\x00" +
						"A  foo -> bar\x00" +
						"MM back\\slashes\x00" +
						"R  this is the new filename\x00R  this is really the old name, does it throw off the parser\x00" +
						"??  leading-space\x00",
				},

				{Command: "git hash-object -- modified.txt copied to new.txt new-file foo -> bar back\\slashes this is the new filename  leading-space", Output: "123456\nc0c0c0\n654321\n758213\nfbfbfb\nbbbbbb\nffffff\ncccccc\n"},
			},
			// pathsToHash: []string{"modified.txt", "copied to", "new.txt", "new-file", "foo -> bar", `back\slashes`, "this is the new filename", " leading-space"},
			result: map[string]string{
				"modified.txt":             "123456",
				"copied to":                "c0c0c0",
				"deleted.txt":              "deleted",
				"new.txt":                  "654321",
				"new-dir/":                 "directory",
				"new-file":                 "758213",
				"foo -> bar":               "fbfbfb",
				`back\slashes`:             "bbbbbb",
				"this is the new filename": "ffffff",
				" leading-space":           "cccccc",
			},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			repository := &Repository{
				Git: &CommandExecutor{
					mu:        new(sync.Mutex),
					cmd:       cmdtest.NewOrdered(t, tt.git),
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
