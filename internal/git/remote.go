package git

import (
	"path/filepath"
	"strings"
)

const (
	remotesFolder     = "lefthook-remotes"
	remotesFolderMode = 0o755
)

// RemoteFolder returns the path to the folder where the remote
// repository is located.
func (r *Repo) RemoteFolder(url string, ref string) string {
	return filepath.Join(
		r.RemotesFolder(),
		RemoteDirectoryName(url, ref),
	)
}

// RemotesFolder returns the path to the lefthook remotes folder.
func (r *Repo) RemotesFolder() string {
	return filepath.Join(r.InfoPath, remotesFolder)
}

func (r *Repo) UpdateRemote(path, ref string) error {
	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := r.Git.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()

	if len(ref) != 0 {
		_, err := git.Cmd([]string{
			"git", "-C", path, "fetch", "--quiet", "--depth", "1",
			"origin", "--", ref,
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

func (r *Repo) CloneRemote(dest, directoryName, url, ref string) error {
	cmdClone := []string{"git", "-C", dest, "clone", "--quiet", "--origin", "origin", "--depth", "1"}
	if len(ref) > 0 {
		cmdClone = append(cmdClone, "--branch", ref)
	}
	cmdClone = append(cmdClone, url, directoryName)

	git := r.Git.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()
	_, err := git.Cmd(cmdClone)
	if err != nil {
		return err
	}

	path := filepath.Join(dest, directoryName)
	if len(ref) != 0 {
		_, err := git.Cmd([]string{
			"git", "-C", path, "fetch", "--quiet", "--depth", "1",
			"origin", "--", ref,
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
	}

	return nil
}

func RemoteDirectoryName(url, ref string) string {
	name := filepath.Base(
		strings.TrimSuffix(url, filepath.Ext(url)),
	)

	if ref != "" {
		name = name + "-" + ref
	}

	return name
}
