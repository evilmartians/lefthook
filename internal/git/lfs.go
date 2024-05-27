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

var lfsHookConsumeStdin = [...]string{
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

// DoesLFSHookConsumeStdin returns whether the LFS hookName will consume Stdin
// meaning it won't be available to following commands.
func DoesLFSHookConsumeStdin(hookName string) bool {
	for _, lfsHookName := range lfsHookConsumeStdin {
		if lfsHookName == hookName {
			return true
		}
	}

	return false
}
