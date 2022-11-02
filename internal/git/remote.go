package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
)

const (
	remotesFolder     = "remotes"
	remotesFolderMode = 0o755
)

func (r *Repository) RemoteFolder(url string) string {
	remotesPath := filepath.Join(r.InfoPath, remotesFolder)

	return filepath.Join(
		remotesPath,
		filepath.Base(
			strings.TrimSuffix(url, filepath.Ext(url)),
		),
	)
}

// SyncRemote clones or pulls the latest changes for a git repository that was
// specified as a remote config repository. If successful, the path to the root
// of the repository will be returned.
func (r *Repository) SyncRemote(url, ref string) error {
	remotesPath := filepath.Join(r.InfoPath, remotesFolder)

	err := r.Fs.MkdirAll(remotesPath, remotesFolderMode)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	remotePath := filepath.Join(
		remotesPath,
		filepath.Base(
			strings.TrimSuffix(url, filepath.Ext(url)),
		),
	)

	_, err = r.Fs.Stat(remotePath)
	if err == nil {
		if err := r.updateRemote(remotePath, ref); err != nil {
			return err
		}

		return nil
	}

	if err := r.cloneRemote(remotesPath, url, ref); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateRemote(path, ref string) error {
	log.Debugf("Updating remote config repository: %s", path)

	cmdFetch := []string{"git", "-C", path, "pull", "--quiet"}
	if len(ref) == 0 {
		cmdFetch = append(cmdFetch, "origin", ref)
	}

	_, err := execGit(strings.Join(cmdFetch, " "))
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) cloneRemote(path, url, ref string) error {
	log.Debugf("Cloning remote config repository: %v", path)

	cmdClone := []string{"git", "-C", path, "clone", "--quiet", "--depth", "1"}
	if len(ref) > 0 {
		cmdClone = append(cmdClone, "--branch", ref)
	}
	cmdClone = append(cmdClone, url)

	_, err := execGit(strings.Join(cmdClone, " "))
	if err != nil {
		return err
	}

	return nil
}
