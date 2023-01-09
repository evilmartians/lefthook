package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

const (
	cmdRootPath    = "git rev-parse --show-toplevel"
	cmdHooksPath   = "git rev-parse --git-path hooks"
	cmdInfoPath    = "git rev-parse --git-path info"
	cmdGitPath     = "git rev-parse --git-dir"
	cmdStagedFiles = "git diff --name-only --cached --diff-filter=ACMR"
	cmdAllFiles    = "git ls-files --cached"
	cmdPushFiles   = "git diff --name-only HEAD @{push} || git diff --name-only HEAD master"
	infoDirMode    = 0o775
	cmdStatusShort = "git status --short"
	cmdCreateStash = "git stash create"
	stashMessage   = "lefthook auto backup"
)

// Repository represents a git repository.
type Repository struct {
	Fs                afero.Fs
	HooksPath         string
	RootPath          string
	GitPath           string
	InfoPath          string
	unstagedPatchPath string
}

// NewRepository returns a Repository or an error, if git repository it not initialized.
func NewRepository(fs afero.Fs) (*Repository, error) {
	rootPath, err := execGit(cmdRootPath)
	if err != nil {
		return nil, err
	}

	hooksPath, err := execGit(cmdHooksPath)
	if err != nil {
		return nil, err
	}
	if exists, _ := afero.DirExists(fs, filepath.Join(rootPath, hooksPath)); exists {
		hooksPath = filepath.Join(rootPath, hooksPath)
	}

	infoPath, err := execGit(cmdInfoPath)
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

	gitPath, err := execGit(cmdGitPath)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(gitPath) {
		gitPath = filepath.Join(rootPath, gitPath)
	}

	return &Repository{
		Fs:                fs,
		HooksPath:         hooksPath,
		RootPath:          rootPath,
		GitPath:           gitPath,
		InfoPath:          infoPath,
		unstagedPatchPath: filepath.Join(infoPath, "lefthook-unstaged.patch"),
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
	return r.FilesByCommand(cmdPushFiles)
}

// PartiallyStagedFiles returns the list of files that have both staged and
// unstaged changes.
// See https://git-scm.com/docs/git-status#_short_format.
func (r *Repository) PartiallyStagedFiles() ([]string, error) {
	lines, err := r.gitLines(cmdStatusShort)
	if err != nil {
		return []string{}, err
	}

	partiallyStaged := make([]string, 0)

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		m1 := line[0] // index
		m2 := line[1] // working tree
		filename := line[3:]
		if m1 != ' ' && m1 != '?' && m2 != ' ' && m2 != '?' && len(filename) > 0 {
			partiallyStaged = append(partiallyStaged, filename)
		}
	}

	return partiallyStaged, nil
}

func (r *Repository) SaveUnstaged(files []string) error {
	_, err := execGitCmd(
		append([]string{
			"git",
			"diff",
			"--binary",          // support binary files
			"--unified=0",       // do not add lines around diff for consistent behaviour
			"--no-color",        // disable colors for consistent behaviour
			"--no-ext-diff",     // disable external diff tools for consistent behaviour
			"--src-prefix=a/",   // force prefix for consistent behaviour
			"--dst-prefix=b/",   // force prefix for consistent behaviour
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
	_, err := execGitCmd(
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

	_, err := execGitCmd(
		"git",
		"apply",
		"-v",
		"--whitespace=nowarn",
		"--recount",
		"--unidiff-zero",
		r.unstagedPatchPath,
	)

	if err == nil {
		err = r.Fs.Remove(r.unstagedPatchPath)
	}

	return err
}

func (r *Repository) StashUnstaged() (string, error) {
	stashHash, err := execGit(cmdCreateStash)
	if err != nil {
		return "", err
	}

	_, err = execGitCmd(
		"git",
		"stash",
		"store",
		"--quiet",
		"--message",
		stashMessage,
		stashHash,
	)
	if err != nil {
		return "", err
	}

	return stashHash, nil
}

func (r *Repository) DropUnstagedStash(hash string) error {
	_, err := execGitCmd(
		"git",
		"stash",
		"drop",
		"--quiet",
		hash,
	)

	return err
}

// FilesByCommand accepts git command and returns its result as a list of filepaths.
func (r *Repository) FilesByCommand(command string) ([]string, error) {

	lines, err := r.gitLines(command)
	if err != nil {
		return nil, err
	}

	return r.extractFiles(lines)
}

func (r *Repository) gitLines(command string) ([]string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		commandArg := strings.Split(command, " ")
		cmd = exec.Command(commandArg[0], commandArg[1:]...)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(outputBytes), "\n"), nil
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
	stat, err := r.Fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}

func execGit(command string) (string, error) {
	args := strings.Split(command, " ")
	return execGitCmd(args...)
}

// execGitCmd executes git command with LEFTHOOK=0 in order
// to prevent calling subsequent lefthook hooks.
func execGitCmd(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "LEFTHOOK=0")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
