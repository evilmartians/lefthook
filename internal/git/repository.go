package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
)

var (
	reHeadBranch     = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)
	reVersion        = regexp.MustCompile(`\d+\.\d+\.\d+`)
	cmdPushFilesBase = []string{"git", "diff", "--name-only", "HEAD", "@{push}"}
	cmdPushFilesHead = []string{"git", "diff", "--name-only", "HEAD"}
	cmdStagedFiles   = []string{"git", "diff", "--name-only", "--cached", "--diff-filter=ACMR"}
	cmdStatusShort   = []string{"git", "status", "--short", "--porcelain"}
	cmdListStash     = []string{"git", "stash", "list"}
	cmdRootPath      = []string{"git", "rev-parse", "--path-format=absolute", "--show-toplevel"}
	cmdHooksPath     = []string{"git", "rev-parse", "--path-format=absolute", "--git-path", "hooks"}
	cmdInfoPath      = []string{"git", "rev-parse", "--path-format=absolute", "--git-path", "info"}
	cmdGitPath       = []string{"git", "rev-parse", "--path-format=absolute", "--git-dir"}
	cmdAllFiles      = []string{"git", "ls-files", "--cached"}
	cmdCreateStash   = []string{"git", "stash", "create"}
	cmdStageFiles    = []string{"git", "add"}
	cmdRemotes       = []string{"git", "branch", "--remotes"}
	cmdHideUnstaged  = []string{"git", "checkout", "--force", "--"}
	cmdEmptyTreeSHA  = []string{"git", "hash-object", "-t", "tree", "/dev/null"}
	cmdGitVersion    = []string{"git", "version"}
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
	emptyTreeSHA      string
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

	rootPath, err := git.Cmd(cmdRootPath)
	if err != nil {
		return nil, err
	}

	hooksPath, err := git.Cmd(cmdHooksPath)
	if err != nil {
		return nil, err
	}

	infoPath, err := git.Cmd(cmdInfoPath)
	if err != nil {
		return nil, err
	}
	infoPath = filepath.Clean(infoPath)

	if exists, _ := afero.DirExists(fs, infoPath); !exists {
		err = fs.Mkdir(infoPath, infoDirMode)
		if err != nil {
			return nil, err
		}
	}

	gitPath, err := git.Cmd(cmdGitPath)
	if err != nil {
		return nil, err
	}

	emptyTreeSHA, err := git.Cmd(cmdEmptyTreeSHA)
	if err != nil {
		log.Debug("Couldn't get empty tree SHA value, not critical")
	}

	git.root = rootPath

	return &Repository{
		Fs:                fs,
		Git:               git,
		HooksPath:         hooksPath,
		RootPath:          rootPath,
		GitPath:           gitPath,
		InfoPath:          infoPath,
		unstagedPatchPath: filepath.Join(infoPath, unstagedPatchName),
		emptyTreeSHA:      emptyTreeSHA,
	}, nil
}

// StagedFiles returns a list of staged files
// or an error if git command fails.
func (r *Repository) StagedFiles() ([]string, error) {
	return r.FilesByCommand(cmdStagedFiles, "")
}

// StagedFiles returns a list of all files in repository
// or an error if git command fails.
func (r *Repository) AllFiles() ([]string, error) {
	return r.FilesByCommand(cmdAllFiles, "")
}

// PushFiles returns a list of files that are ready to be pushed
// or an error if git command fails.
func (r *Repository) PushFiles() ([]string, error) {
	res, err := r.FilesByCommand(cmdPushFilesBase, "")
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
		r.headBranch = r.emptyTreeSHA
	}

	return r.FilesByCommand(append(cmdPushFilesHead, r.headBranch), "")
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repository) PartiallyStagedFiles() ([]string, error) {
	lines, err := r.Git.CmdLines(cmdStatusShort)
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

// FilesByCommand accepts git command and returns its result as a list of filepaths.
func (r *Repository) FilesByCommand(command []string, folder string) ([]string, error) {
	lines, err := r.Git.CmdLinesWithinFolder(command, folder)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines)
}

func (r *Repository) extractFiles(lines []string) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 {
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
