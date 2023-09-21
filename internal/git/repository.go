package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

const (
	cmdRootPath      = "git rev-parse --show-toplevel"
	cmdHooksPath     = "git rev-parse --git-path hooks"
	cmdInfoPath      = "git rev-parse --git-path info"
	cmdGitPath       = "git rev-parse --git-dir"
	cmdStagedFiles   = "git diff --name-only --cached --diff-filter=ACMR"
	cmdAllFiles      = "git ls-files --cached"
	cmdPushFilesBase = "git diff --name-only HEAD @{push}"
	cmdPushFilesHead = "git diff --name-only HEAD %s"
	cmdStatusShort   = "git status --short"
	cmdCreateStash   = "git stash create"
	cmdListStash     = "git stash list"

	stashMessage      = "lefthook auto backup"
	unstagedPatchName = "lefthook-unstaged.patch"
	infoDirMode       = 0o775
	minStatusLen      = 3
)

var headBranchRegexp = regexp.MustCompile(`HEAD -> (?P<name>.*)$`)

// Repository represents a git repository.
type Repository struct {
	Fs                afero.Fs
	Git               Exec
	HooksPath         string
	RootPath          string
	GitPath           string
	InfoPath          string
	unstagedPatchPath string
	headBranch        string
}

// NewRepository returns a Repository or an error, if git repository it not initialized.
func NewRepository(fs afero.Fs, git Exec) (*Repository, error) {
	rootPath, err := git.Cmd(cmdRootPath)
	if err != nil {
		return nil, err
	}

	hooksPath, err := git.Cmd(cmdHooksPath)
	if err != nil {
		return nil, err
	}
	if exists, _ := afero.DirExists(fs, filepath.Join(rootPath, hooksPath)); exists {
		hooksPath = filepath.Join(rootPath, hooksPath)
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
	if !filepath.IsAbs(gitPath) {
		gitPath = filepath.Join(rootPath, gitPath)
	}

	git.SetRootPath(rootPath)

	return &Repository{
		Fs:                fs,
		Git:               git,
		HooksPath:         hooksPath,
		RootPath:          rootPath,
		GitPath:           gitPath,
		InfoPath:          infoPath,
		unstagedPatchPath: filepath.Join(infoPath, unstagedPatchName),
	}, nil
}

// StagedFiles returns a list of staged files
// or an error if git command fails.
func (r *Repository) StagedFiles() ([]string, error) {
	return r.FilesByCommand(cmdStagedFiles)
}

// StagedFiles returns a list of all files in repository
// or an error if git command fails.
func (r *Repository) AllFiles() ([]string, error) {
	return r.FilesByCommand(cmdAllFiles)
}

// PushFiles returns a list of files that are ready to be pushed
// or an error if git command fails.
func (r *Repository) PushFiles() ([]string, error) {
	res, err := r.FilesByCommand(cmdPushFilesBase)
	if err == nil {
		return res, nil
	}

	if len(r.headBranch) == 0 {
		branches, err := r.Git.CmdLines("git branch --remotes")
		if err != nil {
			return nil, err
		}
		for _, branch := range branches {
			if !headBranchRegexp.MatchString(branch) {
				continue
			}

			matches := headBranchRegexp.FindStringSubmatch(branch)
			r.headBranch = matches[headBranchRegexp.SubexpIndex("name")]
			break
		}
	}
	return r.FilesByCommand(fmt.Sprintf(cmdPushFilesHead, r.headBranch))
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
	_, err := r.Git.CmdArgs(
		append([]string{
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
		}, files...)...,
	)

	return err
}

func (r *Repository) HideUnstaged(files []string) error {
	_, err := r.Git.CmdArgs(
		append([]string{
			"git",
			"checkout",
			"--force",
			"--",
		}, files...)...,
	)

	return err
}

func (r *Repository) RestoreUnstaged() error {
	if ok, _ := afero.Exists(r.Fs, r.unstagedPatchPath); !ok {
		return nil
	}

	_, err := r.Git.CmdArgs(
		"git",
		"apply",
		"-v",
		"--whitespace=nowarn",
		"--recount",
		"--unidiff-zero",
		"--allow-empty",
		r.unstagedPatchPath,
	)

	if err == nil {
		err = r.Fs.Remove(r.unstagedPatchPath)
	}

	return err
}

func (r *Repository) StashUnstaged() error {
	stashHash, err := r.Git.Cmd(cmdCreateStash)
	if err != nil {
		return err
	}

	_, err = r.Git.CmdArgs(
		"git",
		"stash",
		"store",
		"--quiet",
		"--message",
		stashMessage,
		stashHash,
	)
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
			_, err := r.Git.CmdArgs(
				"git",
				"stash",
				"drop",
				"--quiet",
				matches[stashID],
			)
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

	_, err := r.Git.CmdArgs(
		append([]string{"git", "add"}, files...)...,
	)

	return err
}

// FilesByCommand accepts git command and returns its result as a list of filepaths.
func (r *Repository) FilesByCommand(command string) ([]string, error) {
	lines, err := r.Git.CmdLines(command)
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
