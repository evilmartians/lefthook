package wrapper

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

var cmdPushFilesBase = []string{
	"git", "diff", "--name-only", "HEAD", "@{push}",
}

var cmdPushFilesHead = []string{
	"git", "diff", "--name-only", // origin/main...HEAD
}

var cmdLsTreeFilesHead = []string{
	"git", "ls-tree", "-r", "--name-only", "HEAD",
}

// PushFiles returns the list of files changed between the local branch
// and the remote branch.
//
// The list of files will be the best effort of:
// 1. Trying to compare current HEAD with @{push} ref
// 2. Trying to compare current HEAD with origin/<current-branch>
// 3. Returning all files known to Git.
func (w *Wrapper) PushFiles() ([]string, error) {
	// Upstream can be unset, do not log the failure as an error
	lines, err := w.cmd.OnlyDebugLogs().cmdLines(cmdPushFilesBase)
	if err == nil { // ignoring error for best effort
		return w.existingFilepaths(lines), nil
	}

	if len(w.cache.headBranch) == 0 {
		w.cache.headBranch = w.resolveHeadBranch()
	}

	if len(w.cache.headBranch) != 0 {
		lines, err = w.cmd.OnlyDebugLogs().cmdLines(append(cmdPushFilesHead, w.cache.headBranch+"...HEAD", "--"))
		if err == nil { // ignoring error for best effort
			return w.existingFilepaths(lines), nil
		}
	}

	// Nothing has been pushed yet or upstream is not set
	return w.Files(cmdLsTreeFilesHead)
}

// PushFilesFromRefs returns the union of files changed by Git's pre-push ref updates.
// Paths belong to the pushed trees, which need not be checked out.
func (w *Wrapper) PushFilesFromRefs(remote, input string) ([]string, error) {
	var files []string
	for line := range strings.SplitSeq(input, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 4 || !validPushOID(fields[1]) || !validPushOID(fields[3]) {
			return nil, fmt.Errorf("invalid pre-push ref update: %q", line)
		}

		localOID, remoteOID := fields[1], fields[3]
		if strings.Trim(localOID, "0") == "" {
			continue
		}

		revisions := []string{remoteOID, localOID}
		if strings.Trim(remoteOID, "0") == "" {
			revisions = []string{"refs/remotes/" + remote + "/HEAD..." + localOID}
		}
		args := append([]string{"git", "diff", "--name-only", "-z", "--diff-filter=d"}, revisions...)
		out, err := w.cmd.OnlyDebugLogs().WithoutTrim().cmd(append(args, "--"))
		if err != nil {
			// The remote base can be unavailable on an initial or forced push.
			// Fall back to the pushed tree, never the checked-out HEAD.
			out, err = w.cmd.OnlyDebugLogs().WithoutTrim().cmd([]string{"git", "ls-tree", "-r", "--name-only", "-z", localOID})
			if err != nil {
				kind, kindErr := w.cmd.OnlyDebugLogs().cmd([]string{"git", "cat-file", "-t", localOID + "^{}"})
				if kindErr == nil && kind == "blob" {
					// A tag pointing to a blob has no changed file paths.
					continue
				}
				return nil, fmt.Errorf("list files for pushed ref %s: %w", fields[0], err)
			}
		}
		for path := range strings.SplitSeq(out, "\x00") {
			if path != "" {
				files = append(files, path)
			}
		}
	}
	slices.Sort(files)
	return slices.Compact(files), nil
}

func validPushOID(oid string) bool {
	if len(oid) != 40 && len(oid) != 64 {
		return false
	}
	_, err := hex.DecodeString(oid)
	return err == nil
}
