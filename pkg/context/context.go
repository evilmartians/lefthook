package context

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func StagedFiles() ([]string, error) {
	return ExecGitCommand("git diff --name-only --cached")
}

func AllFiles() ([]string, error) {
	return ExecGitCommand("git ls-files --cached")
}

func PushFiles() ([]string, error) {
	return ExecGitCommand("git diff --name-only HEAD @{push} || git diff --name-only HEAD master")
}

func ExecGitCommand(command string) ([]string, error) {
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

	return extractFiles(lines)
}

func extractFiles(lines []string) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 {
			continue
		}

		isFile, err := isFile(file)
		if err != nil {
			return nil, err
		}

		if isFile {
			files = append(files, file)
		}
	}

	return files, nil
}

func isFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}
