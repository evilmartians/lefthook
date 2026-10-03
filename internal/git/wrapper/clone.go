package wrapper

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
	if len(args.Ref) > 0 {
		cmdClone = append(cmdClone, "--branch", args.Ref)
	}
	cmdClone = append(cmdClone, args.Url, args.Dest)

	_, err := git.cmd(cmdClone)
	return err
}
