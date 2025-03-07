package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/evilmartians/lefthook/internal/log"
)

const (
	remotesFolder     = "lefthook-remotes"
	remotesFolderMode = 0o755
)

// RemoteFolder returns the path to the folder where the remote
// repository is located.
func (r *Repository) RemoteFolder(url string, ref string) string {
	return filepath.Join(
		r.RemotesFolder(),
		remoteDirectoryName(url, ref),
	)
}

// RemotesFolder returns the path to the lefthook remotes folder.
func (r *Repository) RemotesFolder() string {
	return filepath.Join(r.InfoPath, remotesFolder)
}

// SyncRemote clones or pulls the latest changes for a git repository that was
// specified as a remote config repository. If successful, the path to the root
// of the repository will be returned.
func (r *Repository) SyncRemote(url, ref string, force bool) error {
	remotesPath := r.RemotesFolder()

	err := r.Fs.MkdirAll(remotesPath, remotesFolderMode)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	log.SetName("fetching remotes")
	log.StartSpinner()
	defer log.StopSpinner()
	defer log.UnsetName("fetching remotes")

	directoryName := remoteDirectoryName(url, ref)
	remotePath := filepath.Join(remotesPath, directoryName)

	if force {
		err = r.Fs.RemoveAll(remotePath)
		if err != nil {
			return err
		}
	} else {
		_, err = r.Fs.Stat(remotePath)
		if err == nil {
			return r.updateRemote(remotePath, ref)
		}
	}

	return r.cloneRemote(remotesPath, directoryName, url, ref)
}

func (r *Repository) updateRemote(path, ref string) error {
	log.Debugf("Updating remote config repository: %s", path)

	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := r.Git.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE")

	if len(ref) != 0 {
		_, err := git.Cmd([]string{
			"git", "-C", path, "fetch", "--quiet", "--depth", "1",
			"origin", ref,
		})
		if err != nil {
			return err
		}

		_, err = git.Cmd([]string{
			"git", "-C", path, "checkout", "FETCH_HEAD",
		})
		if err != nil {
			return err
		}
	} else {
		_, err := git.Cmd([]string{"git", "-C", path, "pull", "--quiet"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) cloneRemote(dest, directoryName, url, ref string) error {
	log.Debugf("Cloning remote config repository: %v/%v", dest, directoryName)

	cmdClone := []string{"git", "-C", dest, "clone", "--quiet", "--origin", "origin", "--depth", "1"}
	if len(ref) > 0 {
		cmdClone = append(cmdClone, "--branch", ref)
	}
	cmdClone = append(cmdClone, url, directoryName)

	_, err := r.Git.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").Cmd(cmdClone)
	if err != nil {
		return err
	}

	return nil
}

func remoteDirectoryName(url, ref string) string {
	name := filepath.Base(
		strings.TrimSuffix(url, filepath.Ext(url)),
	)

	if ref != "" {
		name = name + "-" + ref
	}

	return name
}
