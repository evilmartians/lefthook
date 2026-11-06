package wrapper

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

type CloneArgs struct {
	Root string
	Dest string
	Url  string
	Ref  string
}

func (w *Wrapper) Clone(args CloneArgs) error {
	// This is overwriting ENVs for worktrees, otherwise it does not work.
	git := w.cmd.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()

	ref, isSHA, err := w.classifyRef(args.Url, args.Ref)
	if err != nil {
		return err
	}

	if isSHA {
		if _, err = git.cmd([]string{
			"git", "-C", args.Root, "init", "--quiet", args.Dest,
		}); err != nil {
			return err
		}

		_, err = git.cmd([]string{
			"git", "-C", filepath.Join(args.Root, args.Dest),
			"remote", "add", "origin", args.Url,
		})
		return err
	}

	cmdClone := []string{"git", "-C", args.Root, "clone", "--quiet", "--origin", "origin", "--depth", "1"}
	if len(ref) > 0 {
		cmdClone = append(cmdClone, "--branch", ref)
	}
	cmdClone = append(cmdClone, args.Url, args.Dest)

	_, err = git.cmd(cmdClone)
	return err
}

// classifyRef distinguishes a branch or tag from a full commit SHA.
func (w *Wrapper) classifyRef(url, ref string) (string, bool, error) {
	kind := refKindOf(ref)
	if kind == refNamed {
		return ref, false, nil
	}

	if _, ok := w.remoteRefName("", url, ref); ok {
		return ref, false, nil
	}

	if kind == refAbbreviatedSHA {
		return "", false, fmt.Errorf(
			"ref %q looks like an abbreviated commit SHA; use the full 40- or 64-character SHA",
			ref,
		)
	}

	return ref, true, nil
}

// remoteRefName returns the full name of a branch or tag on remote whose name matches ref, if any.
func (w *Wrapper) remoteRefName(root, remote, ref string) (string, bool) {
	git := w.cmd.WithoutEnvs("GIT_DIR", "GIT_INDEX_FILE").OnlyDebugLogs()

	cmd := []string{"git"}
	if len(root) > 0 {
		cmd = append(cmd, "-C", root)
	}
	cmd = append(cmd, "ls-remote", "--refs", "--exit-code", remote, ref)

	out, err := git.cmd(cmd)
	if err != nil {
		return "", false
	}

	line, _, _ := strings.Cut(out, "\n")
	_, name, ok := strings.Cut(line, "\t")

	return name, ok
}

// refKind describes how a configured ref is interpreted.
type refKind int

const (
	refNamed refKind = iota
	refCommitSHA
	refAbbreviatedSHA
)

// refKindOf classifies ref by its hexadecimal shape.
func refKindOf(ref string) refKind {
	if len(ref) == 0 {
		return refNamed
	}

	length := len(ref)
	if length%2 == 1 {
		ref += "0"
	}

	if _, err := hex.DecodeString(ref); err != nil {
		return refNamed
	}

	if length == 40 || length == 64 {
		return refCommitSHA
	}

	return refAbbreviatedSHA
}
