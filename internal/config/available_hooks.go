package config

// ChecksumFileName - the file, which is used just to store the current config checksum version.
const ChecksumFileName = "lefthook.checksum"

// AvailableHooks - list of hooks taken from https://git-scm.com/docs/githooks.
var AvailableHooks = [...]string{
	"pre-applypatch",
	"applypatch-msg",
	"post-applypatch",
	"commit-msg",
	"fsmonitor-watchman",
	"p4-changelist",
	"p4-post-changelist",
	"p4-pre-submit",
	"p4-prepare-changelist",
	"pre-commit",
	"post-commit",
	"pre-receive",
	"proc-receive",
	"post-receive",
	"post-merge",
	"pre-rebase",
	"rebase",
	"update",
	"post-update",
	"post-rewrite",
	"post-checkout",
	"post-index-change",
	"pre-auto-gc",
	"pre-merge-commit",
	"pre-push",
	"prepare-commit-msg",
	"push-to-checkout",
	"reference-transaction",
	"sendemail-validate",
}

func HookAvailable(hook string) bool {
	for _, name := range AvailableHooks {
		if name == hook {
			return true
		}
	}

	return false
}
