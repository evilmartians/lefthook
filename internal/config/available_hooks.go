package config

// ChecksumFileName - the file, which is used just to store the current config checksum version.
const ChecksumFileName = "lefthook.checksum"

// GhostHookName - the hook which logs are not shown and which is used for synchronizing hooks.
const GhostHookName = "prepare-commit-msg"

// AvailableHooks - list of hooks taken from https://git-scm.com/docs/githooks.
var AvailableHooks = [...]string{
	"pre-commit",
	"pre-push",
	"commit-msg",
	"applypatch-msg",
	"fsmonitor-watchman",
	"p4-changelist",
	"p4-post-changelist",
	"p4-pre-submit",
	"p4-prepare-changelist",
	"post-applypatch",
	"post-checkout",
	"post-commit",
	"post-index-change",
	"post-merge",
	"post-receive",
	"post-rewrite",
	"post-update",
	"pre-applypatch",
	"pre-auto-gc",
	"pre-merge-commit",
	"pre-rebase",
	"pre-receive",
	"prepare-commit-msg",
	"proc-receive",
	"push-to-checkout",
	"rebase",
	"reference-transaction",
	"sendemail-validate",
	"update",
}

func HookUsesStagedFiles(hook string) bool {
	return hook == "pre-commit"
}

func HookUsesPushFiles(hook string) bool {
	return hook == "pre-push"
}

func HookAvailable(hook string) bool {
	for _, name := range AvailableHooks {
		if name == hook {
			return true
		}
	}

	return false
}
