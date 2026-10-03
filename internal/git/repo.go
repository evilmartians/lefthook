package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/internal/version"
)

const (
	minGitVersion = "2.31.0"
	infoDirMode   = 0o775
)

var reVersion = regexp.MustCompile(`\d+\.\d+\.(\d+|\w+)`)

type Paths = wrapper.Paths

// Repo is a Git repository controller.
type Repo struct {
	Fs      afero.Fs
	wrapper Wrapper
	Paths   *Paths
	Cache   *Cache
	logger  *logger.Logger
}

type Wrapper interface {
	// Version returns Git executable version
	Version() (string, error)

	// Paths return required paths for lefthook to know about
	Paths() (*wrapper.Paths, error)

	// LocalHooksPath returns configured local hooks path
	LocalHooksPath() string

	// UnsetLocalHooksPath resets the local core.hooksPath
	UnsetLocalHooksPath() error

	// GlobalHooksPath returns configured global hooks path
	GlobalHooksPath() string

	// UnsetLocalHooksPath resets the global core.hooksPath
	UnsetGlobalHooksPath() error

	// AllFiles returns all files visible to Git
	AllFiles() ([]string, error)

	// StagedFiles returns files added with git add
	StagedFiles() ([]string, error)

	// StagedFilesWithDeleted returns same files as StagedFiles including the deleted files
	StagedFilesWithDeleted() ([]string, error)

	// PushFiles returns the files that differ in local branch and the upstream branch
	PushFiles() ([]string, error)

	// FilesByCommandRelative accepts user command and a dir, and returns existing files by its response
	FilesByCommandRelative(string, string) ([]string, error)

	// StatusShort returns short Git status about files
	StatusShort() ([]wrapper.FileStatus, error)

	// Diff returns printable diff of passed files (colored or not)
	Diff([]string, bool) (string, error)

	// SaveUnstagedDiff saves the currently unstaged changes diff for further restoration
	SaveUnstagedDiff([]string) error

	// UnstagedDiffApplicable checks if diff with unstaged changes can be applied
	UnstagedDiffApplicable() bool

	// ApplyUnstagedDiff applies whether the all diff changes or only selected,
	// returns wrapper.ErrNoUnstagedDiff when no diff was saved
	ApplyUnstagedDiff(bool) error

	// StoreStash saves the current tree into a stash for backup
	StoreStash() error

	// DropStash deletes the created stash
	DropStash() error

	// DiscardUnstagedChanges discards the unstaged changes in given files
	DiscardUnstagedChanges([]string) error

	// DiscardUnstagedChanges discards the all unstaged changes in the project
	DiscardAllUnstagedChanges() error

	// StageFiles runs 'git add' on given files
	StageFiles([]string) error

	// HashObjects returns Git hashes for given files
	HashObjects([]string) ([]string, error)

	// State returns the current Git state
	State() wrapper.State

	// Clone fetches the remote Git repo
	Clone(wrapper.CloneArgs) error

	// Pull updates the content of the remote Git repo
	Pull(string) error

	// Fetch updates the content of the remote Git repo
	Fetch(string, string) error
}

// BuildRepo returns a Repo or an error, if git repository it not initialized.
func BuildRepo(
	fs afero.Fs,
	logger *logger.Logger,
) (*Repo, error) {
	wrapper := wrapper.New(fs, system.Cmd, logger)

	gitVersion, err := wrapper.Version()
	if err == nil {
		checkGitVersion(gitVersion, logger)
	}

	paths, err := wrapper.Paths()
	if err != nil {
		return nil, err
	}

	if exists, _ := afero.DirExists(fs, paths.Info); !exists {
		err = fs.Mkdir(paths.Info, infoDirMode)
		if err != nil {
			return nil, err
		}
	}

	return NewRepo(
		fs,
		logger,
		wrapper,
		paths,
	), nil
}

func NewRepo(
	fs afero.Fs,
	logger *logger.Logger,
	wrapper Wrapper,
	paths *Paths,
) *Repo {
	return &Repo{
		Fs:      fs,
		Paths:   paths,
		Cache:   newCache(wrapper),
		logger:  logger,
		wrapper: wrapper,
	}
}

// func (repo *Repo) WithLogger(logger *logger.Logger) *Repo {
// 	repo.logger = logger
// 	// repo.Git.logger = logger // TODO
// 	return repo
// }

func checkGitVersion(strVersion string, logger *logger.Logger) {
	gitVersion := reVersion.FindString(strVersion)
	err := version.Check(minGitVersion, gitVersion)
	if err == nil {
		return
	}

	logger.Debugf("[lefthook] version check warning: %s %s", gitVersion, err)

	if errors.Is(err, version.ErrUncoveredVersion) {
		logger.Warn("Git version is too old. Minimum supported version is " + minGitVersion)
	}
}

func (r *Repo) ResetPaths() error {
	paths, err := r.wrapper.Paths()
	if err != nil {
		return err
	}

	r.Paths = paths

	return nil
}

// StagedFiles returns a list of staged files which exist on file system.
func (r *Repo) StagedFiles() ([]string, error) {
	return r.Cache.stagedFilesOnce()
}

// StagedFilesWithDeleted returns a list of staged files with deleted files.
func (r *Repo) StagedFilesWithDeleted() ([]string, error) {
	return r.Cache.stagedFilesWithDeletedOnce()
}

// AllFiles returns a list of all files in repository.
func (r *Repo) AllFiles() ([]string, error) {
	return r.wrapper.AllFiles()
}

// PushFiles returns a list of files that are ready to be pushed.
func (r *Repo) PushFiles() ([]string, error) {
	return r.wrapper.PushFiles()
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repo) PartiallyStagedFiles() ([]string, error) {
	partiallyStaged := make([]string, 0)

	statuses, err := r.Cache.statusShortOnce()
	if err != nil {
		return nil, err
	}

	for _, status := range statuses {
		if status.Index == ' ' {
			continue
		}
		if status.Index == '?' {
			continue
		}
		if status.Worktree == ' ' {
			continue
		}
		if status.Worktree == '?' {
			continue
		}

		partiallyStaged = append(partiallyStaged, status.Path)
	}

	return partiallyStaged, nil
}

func (r *Repo) SaveUnstagedChanges(files []string) error {
	if err := r.wrapper.SaveUnstagedDiff(files); err != nil {
		return err
	}

	return r.wrapper.StoreStash()
}

func (r *Repo) DiscardUnstagedChanges(files []string) error {
	return r.wrapper.DiscardUnstagedChanges(files)
}

func (r *Repo) DiscardAllUnstagedChanges() error {
	return r.wrapper.DiscardAllUnstagedChanges()
}

// CanRestoreUnstagedChanges checks is a patch with previously unstaged changes
// can be applied to the current worktree.
func (r *Repo) CanRestoreUnstagedChanges() bool {
	return r.wrapper.UnstagedDiffApplicable()
}

// RestoreUnstagedChanges applies the patch with previously unstaged changes.
func (r *Repo) RestoreUnstagedChanges() error {
	return r.restoreUnstagedChanges(false)
}

// RestoreAllUnstagedChanges applies all unstaged changes saved before running hooks.
func (r *Repo) RestoreAllUnstagedChanges() error {
	return r.restoreUnstagedChanges(true)
}

func (r *Repo) restoreUnstagedChanges(all bool) error {
	err := r.wrapper.ApplyUnstagedDiff(all)
	if errors.Is(err, wrapper.ErrNoUnstagedDiff) {
		// Keep the backup if there was nothing to restore
		r.logger.Warn(
			"Saved unstaged changes not found. " +
				"Restore them from the 'lefthook auto backup' stash: git stash list",
		)
		return nil
	}
	if err != nil {
		return err
	}

	if err = r.wrapper.DropStash(); err != nil {
		return fmt.Errorf("failed to remove unstaged files backup: %w", err)
	}

	return nil
}

func (r *Repo) AddFiles(files []string) error {
	return r.wrapper.StageFiles(files)
}

// Changeset returns a map of files and their hashes that are different from the index.
// The hash for a deleted file is "deleted", and "directory" for a directory
// (including submodules and symlinks to directories, which git can't hash).
func (r *Repo) Changeset() (map[string]string, error) {
	changeset := make(map[string]string)
	pathsToHash := make([]string, 0)

	statuses, err := r.wrapper.StatusShort()
	if err != nil {
		return nil, err
	}

	for _, status := range statuses {
		if status.Index == 'D' || status.Worktree == 'D' {
			changeset[status.Path] = "deleted"
			continue
		}
		if strings.HasSuffix(status.Path, "/") {
			changeset[status.Path] = "directory"
			continue
		}
		if isDir, _ := afero.IsDir(r.Fs, filepath.Join(r.Paths.Root, status.Path)); isDir {
			changeset[status.Path] = "directory"
			continue
		}

		pathsToHash = append(pathsToHash, status.Path)
	}

	if len(pathsToHash) == 0 {
		return changeset, nil
	}

	hashes, err := r.wrapper.HashObjects(pathsToHash)
	if err != nil {
		return nil, err
	}

	for i, hash := range hashes {
		changeset[pathsToHash[i]] = hash
	}

	return changeset, nil
}

func (r *Repo) LocalHooksPath() string {
	return r.wrapper.LocalHooksPath()
}

func (r *Repo) UnsetLocalHooksPath() error {
	return r.wrapper.UnsetLocalHooksPath()
}

func (r *Repo) GlobalHooksPath() string {
	return r.wrapper.GlobalHooksPath()
}

func (r *Repo) UnsetGlobalHooksPath() error {
	return r.wrapper.UnsetGlobalHooksPath()
}

func (r *Repo) PrintDiff(files []string) {
	slices.Sort(files)

	diff, err := r.wrapper.Diff(files, !r.logger.NoColors())
	if err != nil {
		r.logger.Warnf("Failed to diff changed files: %s", err)
		return
	}

	r.logger.Warn(diff)
}

// FilesByCommandRelative accepts git command and returns its result as a list of filepaths.
func (r *Repo) FilesByCommandRelative(command string, dir string) ([]string, error) {
	return r.wrapper.FilesByCommandRelative(command, dir)
}

func (r *Repo) State() wrapper.State {
	return r.Cache.stateOnce()
}
