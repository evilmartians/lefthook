package git

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/internal/logger"
)

const (
	minGitVersion        = "2.31.0"
	stashMessage         = "lefthook auto backup"
	unstagedPatchName    = "lefthook-unstaged.patch"
	unstagedAllPatchName = "lefthook-unstaged-all.patch"
	infoDirMode          = 0o775
)

var (
	reHeadBranch       = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)
	reOriginHeadBranch = regexp.MustCompile(`ref: refs/remotes/origin/(?P<name>.*)$`)
	reVersion          = regexp.MustCompile(`\d+\.\d+\.(\d+|\w+)`)
	reStashMessage     = regexp.MustCompile(`^(?P<stash>[^ ]+):\s*` + stashMessage)
)

type Paths struct {
	// Project root path
	Root string

	// .git/hooks/ dir path
	Hooks string

	// .git/ dir path
	Git string

	// .git/info/ dir path
	Info string
}

// Repo is a Git repository controller.
type Repo struct {
	Fs      afero.Fs
	Wrapper Wrapper
	Paths   Paths
	Cache   *Cache
	logger  *logger.Logger
}

type Wrapper interface {
	// Version returns Git executable version
	Version() (string, error)

	// Paths return required paths for lefthook to know about
	Paths() (*wrapper.PathsResult, error)

	// AllFiles returns all files visible to Git
	AllFiles() ([]string, error)

	// StagedFiles returns files added with git add
	StagedFiles() ([]string, error)

	// StagedFilesWithDeleted returns same files as StagedFiles including the deleted onces
	StagedFilesWithDeleted() ([]string, error)

	// PushFiles returns the files that differ in local branch and the upstream branch
	PushFiles() ([]string, error)

	// StatusShort returns short Git status about files
	StatusShort() ([]wrapper.FileStatus, error)
}

// BuildRepo returns a Repo or an error, if git repository it not initialized.
func BuildRepo(
	fs afero.Fs,
	logger *logger.Logger,
) (*Repo, error) {
	wrapper := wrapper.New(fs, logger)

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

	r := &Repo{
		Fs:      fs,
		Wrapper: wrapper,
		Paths: Paths{
			Root:  paths.Root,
			Hooks: paths.Hooks,
			Info:  paths.Info,
			Git:   paths.Git,
		},
		Cache:  newCache(wrapper),
		logger: logger,
	}

	return r, nil
}

func (repo *Repo) WithLogger(logger *logger.Logger) *Repo {
	repo.logger = logger
	// repo.Git.logger = logger // TODO
	return repo
}

func checkGitVersion(version string, logger *logger.Logger) {
	gitVersion := reVersion.FindString(gitVersionOut)
	err := version.Check(minGitVersion, gitVersion)
	if err == nil {
		return
	}

	logger.Debugf("[lefthook] version check warning: %s %s", gitVersion, err)

	if errors.Is(err, version.ErrUncoveredVersion) {
		logger.Warn("Git version is too old. Minimum supported version is " + minGitVersion)
	}
}

// r.unstagedPatchPath = filepath.Join(r.InfoPath, unstagedPatchName)
// r.unstagedAllPatchPath = filepath.Join(r.InfoPath, unstagedAllPatchName)

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
	return r.Wrapper.AllFiles()
}

// PushFiles returns a list of files that are ready to be pushed.
func (r *Repo) PushFiles() ([]string, error) {
	return r.Wrapper.PushFiles()
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repo) PartiallyStagedFiles() ([]string, error) {
	partiallyStaged := make([]string, 0)

	statuses, err := r.statusShortOnce()
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

		partiallyStaged = append(partiallyStaged, path)
	}

	return partiallyStaged, nil
}

func (r *Repo) SaveUnstagedChanges(files []string) error {
	stashHash, err := r.Git.Cmd(cmdCreateStash)
	if err != nil {
		return err
	}

	if err = r.saveUnstaged(files); err != nil {
		return err
	}

	if err = r.saveAllUnstaged(); err != nil {
		return err
	}

	_, err = r.Git.Cmd([]string{
		"git",
		"stash",
		"store",
		"--quiet",
		"--message",
		stashMessage,
		stashHash,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) saveUnstaged(files []string) error {
	_, err := r.Git.BatchedCmd(
		[]string{
			"git",
			"diff",
			"--binary",          // support binary files
			"--unified=0",       // do not add lines around diff for consistent behavior
			"--no-color",        // disable colors for consistent behavior
			"--no-ext-diff",     // disable external diff tools for consistent behavior
			"--src-prefix=a/",   // force prefix for consistent behavior
			"--dst-prefix=b/",   // force prefix for consistent behavior
			"--patch",           // output a patch that can be applied
			"--submodule=short", // always use the default short format for submodules
			"--output",
			r.unstagedPatchPath,
			"--",
		}, files)

	return err
}

func (r *Repo) saveAllUnstaged() error {
	_, err := r.Git.Cmd([]string{
		"git",
		"diff",
		"--binary",
		"--unified=0",
		"--no-color",
		"--no-ext-diff",
		"--src-prefix=a/",
		"--dst-prefix=b/",
		"--patch",
		"--submodule=short",
		"--output",
		r.unstagedAllPatchPath,
		"--",
	})
	if err != nil {
		return fmt.Errorf("failed to save all unstaged changes: %w", err)
	}

	return nil
}

func (r *Repo) RevertUnstagedChanges(files []string) error {
	_, err := r.Git.BatchedCmd(cmdHideUnstaged, files)

	return err
}

func (r *Repo) RevertAllUnstagedChanges() error {
	_, err := r.Git.Cmd(cmdHideAllUnstaged)

	return err
}

// CanRestoreUnstagedChanges checks is a patch with previously unstaged changes
// can be applied to the current worktree.
func (r *Repo) CanRestoreUnstagedChanges() bool {
	if ok, _ := afero.Exists(r.Fs, r.unstagedPatchPath); !ok {
		return true
	}

	stat, err := r.Fs.Stat(r.unstagedPatchPath)
	if err != nil {
		return true
	}

	if stat.Size() == 0 {
		return true
	}

	_, err = r.Git.Cmd([]string{
		"git",
		"apply",
		"-v",
		"--whitespace=nowarn",
		"--recount",
		"--unidiff-zero",
		"--check",
		"--",
		r.unstagedPatchPath,
	})

	return err == nil
}

// RestoreUnstagedChanges applies the patch with previously unstaged changes.
func (r *Repo) RestoreUnstagedChanges() error {
	return r.restoreUnstagedChanges(r.unstagedPatchPath)
}

// RestoreAllUnstagedChanges applies all unstaged changes saved before running hooks.
func (r *Repo) RestoreAllUnstagedChanges() error {
	return r.restoreUnstagedChanges(r.unstagedAllPatchPath)
}

func (r *Repo) restoreUnstagedChanges(patchPath string) error {
	exists, err := afero.Exists(r.Fs, patchPath)
	if err != nil {
		return fmt.Errorf("failed to inspect the patch %s: %w", patchPath, err)
	}
	if !exists {
		return nil
	}

	stat, err := r.Fs.Stat(patchPath)
	if err != nil {
		return err
	}

	if stat.Size() > 0 {
		_, err = r.Git.Cmd([]string{
			"git",
			"apply",
			"-v",
			"--whitespace=nowarn",
			"--recount",
			"--unidiff-zero",
			"--",
			patchPath,
		})
		if err != nil {
			return fmt.Errorf("failed to apply the patch %s: %w", patchPath, err)
		}
	}

	for _, path := range []string{r.unstagedPatchPath, r.unstagedAllPatchPath} {
		exists, existsErr := afero.Exists(r.Fs, path)
		if existsErr != nil {
			return fmt.Errorf("failed to inspect the patch %s: %w", path, existsErr)
		}
		if !exists {
			continue
		}
		if err = r.Fs.Remove(path); err != nil {
			return fmt.Errorf("failed to remove the patch %s: %w", path, err)
		}
	}

	if err = r.dropUnstagedStash(); err != nil {
		return fmt.Errorf("failed to remove unstaged files backup: %w", err)
	}

	return nil
}

func (r *Repo) dropUnstagedStash() error {
	lines, err := r.Git.CmdLines(cmdListStash)
	if err != nil {
		return err
	}

	for i := range lines {
		line := lines[len(lines)-i-1]
		matches := reStashMessage.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		stashID := reStashMessage.SubexpIndex("stash")

		if len(matches[stashID]) > 0 {
			_, err := r.Git.Cmd([]string{
				"git",
				"stash",
				"drop",
				"--quiet",
				"--",
				matches[stashID],
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repo) AddFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	_, err := r.Git.BatchedCmd(cmdStageFiles, files)

	return err
}

// Changeset returns a map of files and their hashes that are different from the index.
// The hash for a deleted file is "deleted", and "directory" for a directory.
func (r *Repo) Changeset() (map[string]string, error) {
	changeset := make(map[string]string)
	pathsToHash := make([]string, 0)

	statuses, err := r.Wrapper.StatusShort()
	if err != nil {
		return nil, err
	}

	for _, status := range statuses {
		if status.Index == 'D' || status.Worktree == 'D' {
			changeset[status.Path] = "deleted"
			return
		}
		if strings.HasSuffix(status.Path, "/") {
			changeset[status.Path] = "directory"
			return
		}

		pathsToHash = append(pathsToHash, status.Path)

	}

	if len(pathsToHash) == 0 {
		return changeset, nil
	}

	out, err := r.Git.BatchedCmd([]string{"git", "hash-object", "--"}, pathsToHash)
	if err != nil {
		return nil, err
	}

	hashes := strings.Split(strings.TrimSpace(out), "\n")
	for i, hash := range hashes {
		changeset[pathsToHash[i]] = hash
	}

	return changeset, nil
}

func (r *Repo) PrintDiff(files []string) {
	slices.Sort(files)

	diffCmd := make([]string, 0, 4) //nolint:mnd // 3 or 4 elements
	diffCmd = append(diffCmd, "git", "diff")
	if !r.logger.NoColors() {
		diffCmd = append(diffCmd, "--color")
	}
	diffCmd = append(diffCmd, "--")
	diff, err := r.Git.BatchedCmd(diffCmd, files)
	if err != nil {
		r.logger.Warnf("Failed to diff changed files: %s", err)
		return
	}

	r.logger.Warn(diff)
}

// FindAllFiles accepts git command and returns its result as a list of filepaths.
func (r *Repo) FindAllFiles(command []string, folder string) ([]string, error) {
	lines, err := r.Git.CmdLinesWithinFolder(command, folder)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines, false)
}

// FindExistingFiles accepts git command and returns its result as a list of filepaths.
func (r *Repo) FindExistingFiles(command []string, folder string) ([]string, error) {
	lines, err := r.Git.CmdLinesRelative(command, folder)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines, true)
}

func (r *Repo) extractFiles(lines []string, checkExistence bool) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 {
			continue
		}

		unescaped, err := strconv.Unquote(file)
		if err == nil {
			file = unescaped
		}

		if !checkExistence {
			files = append(files, file)
			continue
		}

		isFile, err := r.isFile(file)
		if err != nil {
			return nil, err
		}
		if isFile {
			files = append(files, file)
		}
	}

	return files, nil
}

func (r *Repo) isFile(path string) (bool, error) {
	if !strings.HasPrefix(path, r.RootPath) {
		path = filepath.Join(r.RootPath, path)
	}
	stat, err := r.Fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}

func (r *Repo) readOriginHead() string {
	originHead := filepath.Join(r.GitPath, "refs", "remotes", "origin", "HEAD")
	if _, err := r.Fs.Stat(originHead); os.IsNotExist(err) {
		return ""
	}

	file, err := r.Fs.Open(originHead)
	if err != nil {
		return ""
	}
	defer func() {
		if err := file.Close(); err != nil {
			r.logger.Warnf("Could not close %s: %s", originHead, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	_ = scanner.Scan()
	match := reOriginHeadBranch.FindStringSubmatch(scanner.Text())
	if match == nil {
		return ""
	}

	// Return the remote-tracking ref (e.g. "origin/main"), not the bare
	// branch name: the bare name may not exist as a local branch (a clone
	// that never checked out the default branch has no local "main"), which
	// makes the diff in PushFiles fail and the whole hook error out. The
	// remote-tracking ref always exists once origin/HEAD does, and reflects
	// what was actually fetched rather than a possibly stale/absent local
	// branch of the same name.
	return "origin/" + match[reOriginHeadBranch.SubexpIndex("name")]
}
