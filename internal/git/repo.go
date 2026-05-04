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
	"sync"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/logger"
	"github.com/evilmartians/lefthook/v2/internal/system"
	"github.com/evilmartians/lefthook/v2/internal/version"
)

const (
	minGitVersion     = "2.31.0"
	stashMessage      = "lefthook auto backup"
	unstagedPatchName = "lefthook-unstaged.patch"
	infoDirMode       = 0o775
)

var (
	reHeadBranch              = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)
	reOriginHeadBranch        = regexp.MustCompile(`ref: refs/remotes/origin/(?P<name>.*)$`)
	reVersion                 = regexp.MustCompile(`\d+\.\d+\.(\d+|\w+)`)
	reStashMessage            = regexp.MustCompile(`^(?P<stash>[^ ]+):\s*` + stashMessage)
	cmdPushFilesBase          = []string{"git", "diff", "--name-only", "HEAD", "@{push}"}
	cmdPushFilesHead          = []string{"git", "diff", "--name-only", "HEAD"}
	cmdLsTreeFilesHead        = []string{"git", "ls-tree", "-r", "--name-only", "HEAD"}
	cmdStagedFiles            = []string{"git", "diff", "--name-only", "--cached", "--diff-filter=ACMR"}
	cmdStagedFilesWithDeleted = []string{"git", "diff", "--name-only", "--cached", "--diff-filter=ACMRD"}
	cmdStatusShort            = []string{"git", "status", "--short", "--porcelain", "-z"}
	cmdListStash              = []string{"git", "stash", "list"}
	cmdPaths                  = []string{
		"git", "rev-parse", "--path-format=absolute",
		"--show-toplevel",
		"--git-path", "hooks",
		"--git-path", "info",
		"--git-dir",
	}
	cmdAllFiles     = []string{"git", "ls-files", "--cached"}
	cmdCreateStash  = []string{"git", "stash", "create"}
	cmdStageFiles   = []string{"git", "add", "--force", "--"}
	cmdRemotes      = []string{"git", "branch", "--remotes"}
	cmdHideUnstaged = []string{"git", "checkout", "--force", "--"}
	cmdGitVersion   = []string{"git", "version"}
)

// Repo represents a git repository.
type Repo struct {
	Fs        afero.Fs
	Git       *Commander
	Logger    *logger.Logger
	HooksPath string
	RootPath  string
	GitPath   string
	InfoPath  string

	unstagedPatchPath string
	headBranch        string

	stagedFilesOnce            func() ([]string, error)
	stagedFilesWithDeletedOnce func() ([]string, error)
	statusShortOnce            func() ([]string, error)
	stateOnce                  func() State
}

// NewRepo returns a Repo or an error, if git repository it not initialized.
func NewRepo(
	fs afero.Fs,
	logger *logger.Logger,
) (*Repo, error) {
	commander := NewCommander(system.Cmd, logger)
	gitVersionOut, err := commander.Cmd(cmdGitVersion)
	if err == nil {
		gitVersion := reVersion.FindString(gitVersionOut)
		if err = version.Check(minGitVersion, gitVersion); err != nil {
			logger.Debugf("[lefthook] version check warning: %s %s", gitVersion, err)

			if errors.Is(err, version.ErrUncoveredVersion) {
				logger.Warn("Git version is too old. Minimum supported version is " + minGitVersion)
			}
		}
	}

	paths, err := commander.Cmd(cmdPaths)
	if err != nil {
		return nil, err
	}

	pathsSplit := strings.Split(paths, "\n")
	rootPath := pathsSplit[0]
	hooksPath := pathsSplit[1]
	infoPath := filepath.Clean(pathsSplit[2])
	gitPath := pathsSplit[3]

	if exists, _ := afero.DirExists(fs, infoPath); !exists {
		err = fs.Mkdir(infoPath, infoDirMode)
		if err != nil {
			return nil, err
		}
	}

	commander.root = rootPath

	r := &Repo{
		Fs:        fs,
		Git:       commander,
		HooksPath: hooksPath,
		RootPath:  rootPath,
		GitPath:   gitPath,
		InfoPath:  infoPath,
		Logger:    logger,
	}

	// TODO: Rename to something like "Init()"
	r.Setup()

	return r, nil
}

// Precompute runs various Git commands in the background so the results are ready.
// This returns a function which can be used to wait for the result. This should
// be invoked to ensure we're not holding any locks on the Git repository.
// TODO: Rename to something with "cache"
func (r *Repo) Precompute() func() {
	var wg sync.WaitGroup

	wg.Go(func() {
		_, _ = r.stagedFilesOnce()
	})

	wg.Go(func() {
		_, _ = r.stagedFilesWithDeletedOnce()
	})

	wg.Go(func() {
		_, _ = r.statusShortOnce()
	})

	return wg.Wait
}

// Setup must be called after you've constructed a Repository directly.
// It's not necessary to invoke if you've used NewRepository.
//
// This can also be called multiple times to reset the cache.
func (r *Repo) Setup() {
	r.stagedFilesOnce = sync.OnceValues(func() ([]string, error) {
		return r.FindExistingFiles(cmdStagedFiles, "")
	})

	r.stagedFilesWithDeletedOnce = sync.OnceValues(func() ([]string, error) {
		return r.FindAllFiles(cmdStagedFilesWithDeleted, "")
	})

	r.statusShortOnce = sync.OnceValues(func() ([]string, error) {
		return r.statusShort()
	})

	r.stateOnce = sync.OnceValue(func() State {
		return r.state()
	})

	r.unstagedPatchPath = filepath.Join(r.InfoPath, unstagedPatchName)
}

// StagedFiles returns a list of staged files which exist on file system.
func (r *Repo) StagedFiles() ([]string, error) {
	return r.stagedFilesOnce()
}

// StagedFilesWithDeleted returns a list of staged files with deleted files.
func (r *Repo) StagedFilesWithDeleted() ([]string, error) {
	return r.stagedFilesWithDeletedOnce()
}

// AllFiles returns a list of all files in repository.
func (r *Repo) AllFiles() ([]string, error) {
	return r.FindExistingFiles(cmdAllFiles, "")
}

// PushFiles returns a list of files that are ready to be pushed.
func (r *Repo) PushFiles() ([]string, error) {
	// Try with @{push}
	lines, err := r.Git.OnlyDebugLogs().CmdLinesWithinFolder(cmdPushFilesBase, "")
	if err == nil {
		return r.extractFiles(lines, true)
	}

	if len(r.headBranch) == 0 {
		r.headBranch = r.resolveHeadBranch()
	}

	// Nothing has been pushed yet or upstream is not set
	if len(r.headBranch) == 0 {
		return r.FindExistingFiles(cmdLsTreeFilesHead, "")
	}

	return r.FindExistingFiles(append(cmdPushFilesHead, r.headBranch, "--"), "")
}

// resolveHeadBranch determines the upstream head branch.
func (r *Repo) resolveHeadBranch() string {
	if branch := r.readOriginHead(); len(branch) > 0 {
		return branch
	}

	branches, err := r.Git.CmdLines(cmdRemotes)
	if err == nil {
		for _, branch := range branches {
			matches := reHeadBranch.FindStringSubmatch(branch)
			if matches == nil {
				continue
			}
			return matches[reHeadBranch.SubexpIndex("name")]
		}
	}

	return ""
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repo) PartiallyStagedFiles() ([]string, error) {
	partiallyStaged := make([]string, 0)

	lines, err := r.statusShortOnce()
	if err != nil {
		return nil, err
	}

	r.parseStatusShort(lines, func(path string, index, worktree rune) {
		if index != ' ' && index != '?' && worktree != ' ' && worktree != '?' {
			partiallyStaged = append(partiallyStaged, path)
		}
	})

	return partiallyStaged, nil
}

func (r *Repo) SaveUnstaged(files []string) error {
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

func (r *Repo) RevertUnstagedChanges(files []string) error {
	_, err := r.Git.BatchedCmd(cmdHideUnstaged, files)

	return err
}

func (r *Repo) RestoreUnstaged() error {
	if ok, _ := afero.Exists(r.Fs, r.unstagedPatchPath); !ok {
		return nil
	}

	stat, err := r.Fs.Stat(r.unstagedPatchPath)
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		err = r.Fs.Remove(r.unstagedPatchPath)
		if err != nil {
			return fmt.Errorf("couldn't remove the patch %s: %w", r.unstagedPatchPath, err)
		}

		return nil
	}

	_, err = r.Git.Cmd([]string{
		"git",
		"apply",
		"-v",
		"--whitespace=nowarn",
		"--recount",
		"--unidiff-zero",
		"--",
		r.unstagedPatchPath,
	})
	if err != nil {
		return fmt.Errorf("couldn't apply the patch %s: %w", r.unstagedPatchPath, err)
	}

	err = r.Fs.Remove(r.unstagedPatchPath)
	if err != nil {
		return fmt.Errorf("couldn't remove the patch %s: %w", r.unstagedPatchPath, err)
	}

	return nil
}

func (r *Repo) StashUnstaged() error {
	stashHash, err := r.Git.Cmd(cmdCreateStash)
	if err != nil {
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

func (r *Repo) DropUnstagedStash() error {
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

	lines, err := r.statusShort()
	if err != nil {
		return nil, err
	}

	r.parseStatusShort(lines, func(path string, index, worktree rune) {
		if index == 'D' || worktree == 'D' {
			changeset[path] = "deleted"
			return
		}
		if strings.HasSuffix(path, "/") {
			changeset[path] = "directory"
			return
		}

		pathsToHash = append(pathsToHash, path)
	})

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
	if log.Colorized() {
		diffCmd = append(diffCmd, "--color")
	}
	diffCmd = append(diffCmd, "--")
	diff, err := r.Git.BatchedCmd(diffCmd, files)
	if err != nil {
		r.Logger.Warnf("Couldn't diff changed files: %s", err)
		return
	}

	r.Logger.Warn(diff)
}

func (r *Repo) statusShort() ([]string, error) {
	return r.Git.WithoutTrim().CmdLines(cmdStatusShort)
}

// parseStatusShort parses short NUL separated porcelain v1 status output.
// https://git-scm.com/docs/git-status#_short_format
func (r *Repo) parseStatusShort(lines []string, cb func(path string, index, worktree rune)) {
	output := strings.Join(lines, "") // there should be only one line with -z
	skip := false
	for item := range strings.SplitSeq(output, "\x00") {
		if skip {
			skip = false
			continue
		}
		rs := []rune(item)
		if len(rs) < 4 || rs[2] != ' ' { // two status characters, space, and a filename
			continue
		}
		if slices.ContainsFunc(rs[0:2], func(r rune) bool {
			return r == 'C' || r == 'R'
		}) {
			// Next item after a Copy or Rename one is expected to be the old name, which we ignore
			skip = true
		}
		cb(string(rs[3:]), rs[0], rs[1])
	}
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
	lines, err := r.Git.CmdLinesWithinFolder(command, folder)
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
			r.Logger.Warnf("Could not close %s: %s", originHead, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	_ = scanner.Scan()
	match := reOriginHeadBranch.FindStringSubmatch(scanner.Text())
	if match == nil {
		return ""
	}

	return match[reHeadBranch.SubexpIndex("name")]
}
