package git

import (
	"os/exec"
)

const (
	LFSRequiredFile = ".lfs-required"
	LFSConfigFile   = ".lfsconfig"
)

var lfsHooks = map[string]struct{}{
	"post-checkout": {},
	"post-commit":   {},
	"post-merge":    {},
	"pre-push":      {},
}

// IsLFSAvailable returns 'true' if git-lfs is installed.
func IsLFSAvailable() bool {
	_, err := exec.LookPath("git-lfs")

	return err == nil
}

// IsLFSHook returns whether the hookName is supported by Git LFS.
func IsLFSHook(hookName string) bool {
	_, ok := lfsHooks[hookName]
	return ok
}
