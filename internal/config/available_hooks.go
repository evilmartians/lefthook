package config

// ChecksumFileName - the file, which is used just to store the current config checksum version.
const ChecksumFileName = "lefthook.checksum"

// GhostHookName - the hook which logs are not shown and which is used for synchronizing hooks.
const GhostHookName = "prepare-commit-msg"

// AvailableHooks - list of hooks taken from https://git-scm.com/docs/githooks.
// Keep the order of the hooks same here for easy syncing.
var AvailableHooks = map[string]struct{}{
	"applypatch-msg":        {},
	"pre-applypatch":        {},
	"post-applypatch":       {},
	"pre-commit":            {},
	"pre-merge-commit":      {},
	"prepare-commit-msg":    {},
	"commit-msg":            {},
	"post-commit":           {},
	"pre-rebase":            {},
	"post-checkout":         {},
	"post-merge":            {},
	"pre-push":              {},
	"pre-receive":           {},
	"update":                {},
	"proc-receive":          {},
	"post-receive":          {},
	"post-update":           {},
	"reference-transaction": {},
	"push-to-checkout":      {},
	"pre-auto-gc":           {},
	"post-rewrite":          {},
	"sendemail-validate":    {},
	"fsmonitor-watchman":    {},
	"p4-changelist":         {},
	"p4-prepare-changelist": {},
	"p4-post-changelist":    {},
	"p4-pre-submit":         {},
	"post-index-change":     {},
}

func HookUsesStagedFiles(hook string) bool {
	return hook == "pre-commit"
}

func HookUsesPushFiles(hook string) bool {
	return hook == "pre-push"
}

func KnownHook(hook string) bool {
	_, ok := AvailableHooks[hook]
	return ok
}
