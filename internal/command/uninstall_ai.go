package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// uninstallAIHooks removes lefthook-managed entries from every known provider
// settings file. User-authored entries are preserved; a file that contained only
// lefthook-managed entries is removed entirely.
func (l *Lefthook) uninstallAIHooks() error {
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
		{
			path:  filepath.Join(l.repo.RootPath, copilotHooksDir, copilotHooksFile),
			strip: stripCopilotLefthookEntries,
		},
	}

	for _, p := range paths {
		if err := l.removeAIHookEntries(p.path, p.strip); err != nil {
			return err
		}
	}

	return nil
}

// removeAIHookEntries reads a provider settings file, strips lefthook-managed
// entries, and writes the result back. If nothing remains, the file is removed.
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
		// Cursor and Copilot hooks files only contain version + hooks; drop version when hooks are gone.
		delete(existing, "version")
	}

	if len(existing) == 0 {
		if err = l.fs.Remove(path); err != nil {
			return fmt.Errorf("could not remove %s: %w", path, err)
		}
		l.logger.Debugf("%s removed", path)
		return nil
	}

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal %s: %w", path, err)
	}
	out = append(out, '\n')

	if err = afero.WriteFile(l.fs, path, out, checksumFileMode); err != nil {
		return err
	}

	l.logger.Debugf("%s updated", path)
	return nil
}
