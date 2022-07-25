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
	cmdStagedFiles = "git diff --name-only --cached"
	cmdAllFiles    = "git ls-files --cached"
	cmdPushFiles   = "git diff --name-only HEAD @{push} || git diff --name-only HEAD master"
)

// Repository represents a git repository.
type Repository struct {
	Fs        afero.Fs
	HooksPath string
	RootPath  string
	GitPath   string
	InfoPath  string
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
	if exists, _ := afero.DirExists(fs, filepath.Join(rootPath, infoPath)); exists {
		infoPath = filepath.Join(rootPath, infoPath)
	}

	gitPath, err := execGit(cmdGitPath)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(gitPath) {
		gitPath = filepath.Join(rootPath, gitPath)
	}

	return &Repository{
		Fs:        fs,
		HooksPath: hooksPath,
		RootPath:  rootPath,
		GitPath:   gitPath,
		InfoPath:  infoPath,
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

// FilesByCommand accepts git command and returns its result as a list of filepaths.
func (r *Repository) FilesByCommand(command string) ([]string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		commandArg := strings.Split(command, " ")
		cmd = exec.Command(commandArg[0], commandArg[1:]...)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(string(outputBytes), "\n")

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
	cmd := exec.Command(args[0], args[1:]...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
