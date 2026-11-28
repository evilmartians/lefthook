package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// uninstallAIHooks removes lefthook-managed entries from Claude/Codex/Cursor/
// Antigravity settings and removes the lefthook-managed Copilot file entirely.
func (l *Lefthook) uninstallAIHooks() error {
	var firstErr error

	paths := []struct {
		path  string
		strip func(map[string]any) map[string]any
	}{
		{
			path:  filepath.Join(l.repo.Paths.Root, claudeSettingsDir, claudeSettingsFile),
			strip: stripLefthookEntries,
		},
		{
			path:  filepath.Join(l.repo.Paths.Root, codexHooksDir, codexHooksFile),
			strip: stripLefthookEntries,
		},
		{
			path:  filepath.Join(l.repo.Paths.Root, cursorHooksDir, cursorHooksFile),
			strip: stripCursorLefthookEntries,
		},
	}

	for _, p := range paths {
		if err := l.removeAIHookEntries(p.path, p.strip); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if err := l.removeAntigravityHookEntries(filepath.Join(l.repo.Paths.Root, antigravityHooksDir, antigravityHooksFile)); err != nil && firstErr == nil {
		firstErr = err
	}

	if err := l.removeAIHookFile(filepath.Join(l.repo.Paths.Root, copilotHooksDir, copilotHooksFile)); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

func (l *Lefthook) removeAIHookFile(path string) error {
	err := l.fs.Remove(path)
	switch {
	case err == nil:
		l.logger.Debugf("%s removed", path)
		return nil
	case errors.Is(err, os.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("could not remove %s: %w", path, err)
	}
}

func (l *Lefthook) removeAIHookEntries(path string, strip func(map[string]any) map[string]any) error {
	existing, ok, err := l.readJSONFile(path)
	if err != nil || !ok {
		return err
	}

	mergedHooks := strip(existing)
	if len(mergedHooks) > 0 {
		existing["hooks"] = mergedHooks
	} else {
		delete(existing, "hooks")
		delete(existing, "version")
	}

	if len(existing) == 0 {
		return l.removeAIHookFile(path)
	}

	return l.writeJSONFile(path, existing)
}

func (l *Lefthook) removeAntigravityHookEntries(path string) error {
	existing, ok, err := l.readJSONFile(path)
	if err != nil || !ok {
		return err
	}

	if _, hasLefthook := existing[antigravityHooksKey]; !hasLefthook {
		return nil
	}

	delete(existing, antigravityHooksKey)

	if len(existing) > 0 {
		return l.writeJSONFile(path, existing)
	}

	if err = l.removeAIHookFile(path); err != nil {
		return err
	}

	return l.removeEmptyDir(filepath.Dir(path))
}

func (l *Lefthook) removeEmptyDir(dir string) error {
	entries, err := afero.ReadDir(l.fs, dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", dir, err)
	}
	if len(entries) > 0 {
		return nil
	}

	if err = l.fs.Remove(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("could not remove directory %s: %w", dir, err)
	}

	l.logger.Debugf("%s removed", dir)
	return nil
}
