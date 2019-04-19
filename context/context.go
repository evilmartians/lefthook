package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func StagedFiles() ([]string, error) {
	return ExecGitCommand("git diff --name-only --cached")
}

func AllFiles() ([]string, error) {
	return ExecGitCommand("git ls-files --cached")
}

func ExecGitCommand(command string) ([]string, error) {
	commandArg := strings.Split(command, " ")
	cmd := exec.Command(commandArg[0], commandArg[1:]...)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(string(outputBytes), "\n")

	return ExtractFiles(lines)
}

func FilterByExt(files []string, ext string) []string {
	filtred := make([]string, 0)

	for _, f := range files {
		if filepath.Ext(f) == ext {
			filtred = append(filtred, f)
		}
	}
	return filtred
}

func IsFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return !stat.IsDir(), nil
}

func IsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return stat.IsDir(), nil
}

func ExtractFiles(lines []string) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 {
			continue
		}

		isFile, err := IsFile(file)
		if err != nil {
			return nil, err
		}

		if isFile {
			files = append(files, file)
		}
	}

	return files, nil
}
