package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// uninstallAIHooks removes lefthook-managed entries from Claude/Codex/Cursor
// settings and removes the lefthook-managed Copilot file entirely.
func (l *Lefthook) uninstallAIHooks() error {
	var firstErr error

	paths := []struct {
		path  string
		strip func(map[string]any) map[string]any
	}{
		{
			path:  filepath.Join(l.repo.RootPath, claudeSettingsDir, claudeSettingsFile),
			strip: stripLefthookEntries,
		},
		{
			path:  filepath.Join(l.repo.RootPath, codexHooksDir, codexHooksFile),
			strip: stripLefthookEntries,
		},
		{
			path:  filepath.Join(l.repo.RootPath, cursorHooksDir, cursorHooksFile),
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

	if err := l.removeAIHookFile(filepath.Join(l.repo.RootPath, copilotHooksDir, copilotHooksFile)); err != nil && firstErr == nil {
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
	data, err := afero.ReadFile(l.fs, path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not read %s: %w", path, err)
	}
	if len(data) == 0 {
		return nil
	}

	existing := make(map[string]any)
	if err = json.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("could not parse %s: %w", path, err)
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
