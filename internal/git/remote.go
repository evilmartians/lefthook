package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
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
// a remote config repository. If successful, the path to the root of the repository will be
// returned.
func InitRemote(fs afero.Fs, url, rev string) (string, error) {
	err := fs.MkdirAll(defaultRemotePath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}

	root := getRemoteDir(url)

	_, err = fs.Stat(root)
	if err == nil {
		if err := updateRemote(fs, root, url, rev); err != nil {
			return "", err
		}
		return root, nil
	}

	if err := cloneRemote(fs, root, url, rev); err != nil {
		return "", err
	}
	return root, nil
}

func updateRemote(fs afero.Fs, root, url, rev string) error {
	log.Debugf("Updating remote config repository: %v", root)
	cmdFetch := []string{"git", "-C", root, "pull", "-q"}
	if len(rev) == 0 {
		cmdFetch = append(cmdFetch, "origin", rev)
	}
	_, err := execGit(strings.Join(cmdFetch, " "))
	if err != nil {
		return err
	}
	return nil
}

func cloneRemote(fs afero.Fs, root, url, rev string) error {
	log.Debugf("Cloning remote config repository: %v", root)
	cmdClone := []string{"git", "-C", defaultRemotePath, "clone", "-q"}
	if len(rev) > 0 {
		cmdClone = append(cmdClone, "-b", rev)
	}
	cmdClone = append(cmdClone, url)
	_, err := execGit(strings.Join(cmdClone, " "))
	if err != nil {
		return err
	}
	return nil
}

func getRemoteDir(url string) string {
	// Removes any suffix that might have been used in the url like '.git'.
	trimmedURL := strings.TrimSuffix(url, filepath.Ext(url))
	return filepath.Join(defaultRemotePath, filepath.Base(trimmedURL))
}
