package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

const (
	minGitVersion     = "2.31.0"
	stashMessage      = "lefthook auto backup"
	unstagedPatchName = "lefthook-unstaged.patch"
	infoDirMode       = 0o775
	minStatusLen      = 3

	// The result of `git hash-object -t tree /dev/null`.
	emptyTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

var (
	reHeadBranch              = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)
	reVersion                 = regexp.MustCompile(`\d+\.\d+\.\d+`)
	cmdPushFilesBase          = []string{"git", "diff", "--name-only", "HEAD", "@{push}"}
	cmdPushFilesHead          = []string{"git", "diff", "--name-only", "HEAD"}
	cmdStagedFiles            = []string{"git", "diff", "--name-only", "--cached", "--diff-filter=ACMR"}
	cmdStagedFilesWithDeleted = []string{"git", "diff", "--name-only", "--cached", "--diff-filter=ACMRD"}
	cmdStatusShort            = []string{"git", "status", "--short", "--porcelain"}
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
	cmdStageFiles   = []string{"git", "add"}
	cmdRemotes      = []string{"git", "branch", "--remotes"}
	cmdHideUnstaged = []string{"git", "checkout", "--force", "--"}
	cmdGitVersion   = []string{"git", "version"}
)

// Repository represents a git repository.
type Repository struct {
	Fs                afero.Fs
	Git               *CommandExecutor
	HooksPath         string
	RootPath          string
	GitPath           string
	InfoPath          string
	unstagedPatchPath string
	headBranch        string

	stagedFilesOnce            func() ([]string, error)
	stagedFilesWithDeletedOnce func() ([]string, error)
	statusShortOnce            func() ([]string, error)
}

// NewRepository returns a Repository or an error, if git repository it not initialized.
func NewRepository(fs afero.Fs, git *CommandExecutor) (*Repository, error) {
	gitVersionOut, err := git.Cmd(cmdGitVersion)
	if err == nil {
		gitVersion := reVersion.FindString(gitVersionOut)
		if err = version.Check(minGitVersion, gitVersion); err != nil {
			log.Debugf("[lefthook] version check warning: %s %s", gitVersion, err)

			if errors.Is(err, version.ErrUncoveredVersion) {
				log.Warn("Git version is too old. Minimum supported version is " + minGitVersion)
			}
		}
	}

	paths, err := git.Cmd(cmdPaths)
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

	git.root = rootPath

	r := &Repository{
		Fs:                fs,
		Git:               git,
		HooksPath:         hooksPath,
		RootPath:          rootPath,
		GitPath:           gitPath,
		InfoPath:          infoPath,
		unstagedPatchPath: filepath.Join(infoPath, unstagedPatchName),
	}

	r.Setup()

	return r, nil
}

// Precompute runs various Git commands in the background so the results are ready.
// This returns a function which can be used to wait for the result. This should
// be invoked to ensure we're not holding any locks on the Git repository.
func (r *Repository) Precompute() func() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = r.stagedFilesOnce()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = r.stagedFilesWithDeletedOnce()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = r.statusShortOnce()
	}()

	return wg.Wait
}

// Setup must be called after you've constructed a Repository directly.
// It's not necessary to invoke if you've used NewRepository.
//
// This can also be called multiple times to reset the cache.
func (r *Repository) Setup() {
	r.stagedFilesOnce = sync.OnceValues(func() ([]string, error) {
		return r.FindExistingFiles(cmdStagedFiles, "")
	})

	r.stagedFilesWithDeletedOnce = sync.OnceValues(func() ([]string, error) {
		return r.FindAllFiles(cmdStagedFilesWithDeleted, "")
	})

	r.statusShortOnce = sync.OnceValues(func() ([]string, error) {
		return r.Git.CmdLines(cmdStatusShort)
	})
}

// StagedFiles returns a list of staged files which exist on file system.
func (r *Repository) StagedFiles() ([]string, error) {
	return r.stagedFilesOnce()
}

// StagedFilesWithDeleted returns a list of staged files with deleted files.
func (r *Repository) StagedFilesWithDeleted() ([]string, error) {
	return r.stagedFilesWithDeletedOnce()
}

// StagedFiles returns a list of all files in repository.
func (r *Repository) AllFiles() ([]string, error) {
	return r.FindExistingFiles(cmdAllFiles, "")
}

// PushFiles returns a list of files that are ready to be pushed.
func (r *Repository) PushFiles() ([]string, error) {
	res, err := r.FindExistingFiles(cmdPushFilesBase, "")
	if err == nil {
		return res, nil
	}

	if len(r.headBranch) == 0 {
		branches, err := r.Git.CmdLines(cmdRemotes)
		if err != nil {
			return nil, err
		}

		for _, branch := range branches {
			if !reHeadBranch.MatchString(branch) {
				continue
			}

			matches := reHeadBranch.FindStringSubmatch(branch)
			r.headBranch = matches[reHeadBranch.SubexpIndex("name")]
			break
		}
	}

	// Nothing has been pushed yet or upstream is not set
	if len(r.headBranch) == 0 {
		r.headBranch = emptyTreeSHA
	}

	return r.FindExistingFiles(append(cmdPushFilesHead, r.headBranch), "")
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repository) PartiallyStagedFiles() ([]string, error) {
	lines, err := r.statusShortOnce()
	if err != nil {
		return []string{}, err
	}

	partiallyStaged := make([]string, 0)

	for _, line := range lines {
		if len(line) < minStatusLen {
			continue
		}

		index := line[0]
		workingTree := line[1]

		filename := line[3:]
		idx := strings.Index(filename, "->")
		if idx != -1 {
			filename = filename[idx+3:]
		}

		if index != ' ' && index != '?' && workingTree != ' ' && workingTree != '?' && len(filename) > 0 {
			partiallyStaged = append(partiallyStaged, filename)
		}
	}

	return partiallyStaged, nil
}

func (r *Repository) SaveUnstaged(files []string) error {
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

func (r *Repository) HideUnstaged(files []string) error {
	_, err := r.Git.BatchedCmd(cmdHideUnstaged, files)

	return err
}

func (r *Repository) RestoreUnstaged() error {
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

func (r *Repository) StashUnstaged() error {
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

func (r *Repository) DropUnstagedStash() error {
	lines, err := r.Git.CmdLines(cmdListStash)
	if err != nil {
		return err
	}

	stashRegexp := regexp.MustCompile(`^(?P<stash>[^ ]+):\s*` + stashMessage)
	for i := range lines {
		line := lines[len(lines)-i-1]
		matches := stashRegexp.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		stashID := stashRegexp.SubexpIndex("stash")

		if len(matches[stashID]) > 0 {
			_, err := r.Git.Cmd([]string{
				"git",
				"stash",
				"drop",
				"--quiet",
				matches[stashID],
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) AddFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	_, err := r.Git.BatchedCmd(cmdStageFiles, files)

	return err
}

// FindAllFiles accepts git command and returns its result as a list of filepaths.
func (r *Repository) FindAllFiles(command []string, folder string) ([]string, error) {
	lines, err := r.Git.CmdLinesWithinFolder(command, folder)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines, false)
}

// FindExistingFiles accepts git command and returns its result as a list of filepaths.
func (r *Repository) FindExistingFiles(command []string, folder string) ([]string, error) {
	lines, err := r.Git.CmdLinesWithinFolder(command, folder)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines, true)
}

func (r *Repository) extractFiles(lines []string, checkExistence bool) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 {
			continue
		}

		unescaped, err := strconv.Unquote(file)
		if err == nil {
			file = unescaped
		} else {
			log.Debug("[lefthook] couldn't unquote "+file, err)
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

func (r *Repository) isFile(path string) (bool, error) {
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
