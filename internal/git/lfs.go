package git

import (
	"os/exec"
)

const (
	LFSRequiredFile = ".lfs-required"
	LFSConfigFile   = ".lfsconfig"
)

var lfsHooks = [...]string{
	"post-checkout",
	"post-commit",
	"post-merge",
	"pre-push",
}

// IsLFSAvailable returns 'true' if git-lfs is installed.
func IsLFSAvailable() bool {
	_, err := exec.LookPath("git-lfs")

	return err == nil
}

// IsLFSHook returns whether the hookName is supported by Git LFS.
func IsLFSHook(hookName string) bool {
	for _, lfsHookName := range lfsHooks {
		if lfsHookName == hookName {
			return true
		}
	}

	return false
}
