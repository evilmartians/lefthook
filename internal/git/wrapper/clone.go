package wrapper

import "encoding/hex"

type CloneArgs struct {
	Root string
	Dest string
	Url  string
	Ref  string
}

func (w *Wrapper) Clone(args CloneArgs) error {
	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := w.cmd.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()

	cmdClone := []string{"git", "-C", args.Root, "clone", "--quiet", "--origin", "origin", "--depth", "1"}
	// `git clone --branch` only accepts branch and tag names, so passing a
	// commit SHA there makes the clone fail. For a SHA, clone the default
	// branch and let Fetch() fetch and check out the requested commit.
	if len(args.Ref) > 0 && !isCommitSHA(args.Ref) {
		cmdClone = append(cmdClone, "--branch", args.Ref)
	}
	cmdClone = append(cmdClone, args.Url, args.Dest)

	_, err := git.cmd(cmdClone)
	return err
}

// isCommitSHA reports whether ref is a full commit hash: 40 hex characters for
// SHA-1 repositories or 64 for SHA-256 repositories. Abbreviated hashes are not
// supported because Git cannot fetch them by object name from a remote.
func isCommitSHA(ref string) bool {
	if len(ref) != 40 && len(ref) != 64 {
		return false
	}

	_, err := hex.DecodeString(ref)
	return err == nil
}
