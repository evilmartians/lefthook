package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var defaultRemotePath = mustGetDefaultRemotesDir()

// mustGetDefaultRemotesDir returns the default directory for the lefthook remotes.
func mustGetDefaultRemotesDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, ".lefthook-remotes")
}

// InitRemote clones or pulls the latest changes for a git repository that was specified as
// a remote configuration. If successful, the path to the root of the repository will be
// returned.
func InitRemote(fs afero.Fs, url, rev string) (string, error) {
	root := getRemoteDir(url)

	_, err := fs.Stat(root)
	if err == nil {
		cmdFetch := strings.Join([]string{"git", "-C", root, "pull", "-q"}, " ")
		_, err = execGit(cmdFetch)
		if err != nil {
			return "", err
		}

		return root, nil
	}

	err = fs.MkdirAll(defaultRemotePath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}

	cmdClone := []string{"git", "-C", defaultRemotePath, "clone", "-q"}
	if len(rev) > 0 {
		cmdClone = append(cmdClone, "-b", rev)
	}
	cmdClone = append(cmdClone, url)
	_, err = execGit(strings.Join(cmdClone, " "))
	if err != nil {
		return "", err
	}

	return root, nil
}

func getRemoteDir(url string) string {
	// Removes any suffix that might have been used in the url like '.git'.
	trimmedURL := strings.TrimSuffix(url, filepath.Ext(url))
	return filepath.Join(defaultRemotePath, filepath.Base(trimmedURL))
}
