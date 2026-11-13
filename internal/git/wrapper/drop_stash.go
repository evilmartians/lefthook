package wrapper

func (w *Wrapper) DropStash() error {
	hash := w.cache.backupStashHash
	if len(hash) == 0 {
		return nil
	}

	// Deletes the ref only if it still points to this backup.
	_, err := w.cmd.cmd([]string{
		"git",
		"update-ref",
		"-d",
		backupRefPrefix + hash,
		hash,
	})

	return err
}
