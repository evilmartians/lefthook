//go:build windows
// +build windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const sh = "sh"
const defaultShPath = `C:\Program Files\Git\bin\sh.exe`

var fullPath = sync.OnceValues(func() (string, error) {
	if _, err := os.Stat(defaultShPath); err == nil {
		return defaultShPath, nil
	}

	shPath, _ := exec.LookPath("sh")
	if len(shPath) > 0 {
		return shPath, nil
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}

	shPath = filepath.Join(gitPath, "..", "..", "bin", "sh.exe")
	if _, err := os.Stat(shPath); err != nil {
		return "", err
	}

	return shPath, nil
})

// Sh returns the path to a shell or an error if it can't find `sh` executable.
func Sh() (string, error) {
	// In case Git runs lefthook from hooks.
	// Git hooks always setup GIT_INDEX env variable so here we check if we are in
	// a Git hook and can use `sh` without specifying the full path. This should cover most use cases.
	if len(os.Getenv("GIT_INDEX_FILE")) != 0 {
		return sh, nil
	}

	// In case you call `lefthook run ...` from the terminal
	shPath, err := fullPath()
	if err != nil {
		return "", fmt.Errorf("`sh` lookup failed: %w", err)
	}

	return shPath, nil
}
