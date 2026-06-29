package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/config"
)

const (
	claudeSettingsDir  = ".claude"
	claudeSettingsFile = "settings.json"

	codexHooksDir  = ".codex"
	codexHooksFile = "hooks.json"

	cursorHooksDir     = ".cursor"
	cursorHooksFile    = "hooks.json"
	cursorHooksVersion = 1

	lefthookRunPrefix = "lefthook run "
)

var errAIHooksMisconfigured = errors.New("ai hooks misconfigured")

func checkAIHookReferences(provider string, events map[string]string, hooks map[string]*config.Hook) []string {
	var missing []string
	for event, hookName := range events {
		if _, ok := hooks[hookName]; !ok {
			missing = append(missing, fmt.Sprintf("ai.%s.%s -> %q", provider, event, hookName))
		}
	}
	return missing
}

// validateAIHooks ensures every hook name referenced under the ai: key points to
// a hook that is actually defined in the config. This catches typos early instead
// of silently writing a `lefthook run <typo>` command that fails when the agent
// event fires.
func (l *Lefthook) validateAIHooks(ai *config.AI, hooks map[string]*config.Hook) error {
	missing := checkAIHookReferences("claude", ai.Claude, hooks)
	missing = append(missing, checkAIHookReferences("codex", ai.Codex, hooks)...)
	missing = append(missing, checkAIHookReferences("cursor", ai.Cursor, hooks)...)

	if len(missing) > 0 {
		slices.Sort(missing)
		for _, msg := range missing {
			l.logger.Errorf("%s", msg)
		}
		return errAIHooksMisconfigured
	}

	return nil
}

// installAIHooks generates (or merges into) provider-specific settings files
// for each configured LLM provider under the ai: key.
func (l *Lefthook) installAIHooks(cfg *config.AI) error {
	if len(cfg.Claude) > 0 {
		path := filepath.Join(l.repo.RootPath, claudeSettingsDir, claudeSettingsFile)
		if err := l.writeAIHookFile(path, cfg.Claude); err != nil {
			return fmt.Errorf("could not write Claude settings: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(cfg.Codex) > 0 {
		path := filepath.Join(l.repo.RootPath, codexHooksDir, codexHooksFile)
		if err := l.writeAIHookFile(path, cfg.Codex); err != nil {
			return fmt.Errorf("could not write Codex hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(cfg.Cursor) > 0 {
		path := filepath.Join(l.repo.RootPath, cursorHooksDir, cursorHooksFile)
		if err := l.writeCursorHookFile(path, cfg.Cursor); err != nil {
			return fmt.Errorf("could not write Cursor hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	return nil
}

// writeAIHookFile reads an existing JSON settings file (if present), removes
// any lefthook-managed entries (commands starting with "lefthook run "), merges
// in the new entries derived from events, and writes the result back.
//
// Both .claude/settings.json and .codex/hooks.json share the same top-level
// structure:
//
//	{
//	  "hooks": {
//	    "<EventName>": [
//	      { "hooks": [{ "type": "command", "command": "lefthook run <hook>" }] }
//	    ]
//	  }
//	}
func (l *Lefthook) writeAIHookFile(path string, events map[string]string) error {
	existing := make(map[string]any)

	data, err := afero.ReadFile(l.fs, path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not read %s: %w", path, err)
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("could not parse %s: %w", path, err)
		}
	}

	mergedHooks := stripLefthookEntries(existing)
	for event, hookName := range events {
		entry := buildHookEntry(hookName)
		if current, ok := mergedHooks[event]; ok {
			if arr, ok := current.([]any); ok {
				mergedHooks[event] = append(arr, entry)
				continue
			}
		}
		mergedHooks[event] = []any{entry}
	}

	existing["hooks"] = mergedHooks

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal %s: %w", path, err)
	}
	out = append(out, '\n')

	dir := filepath.Dir(path)
	if mkErr := l.fs.MkdirAll(dir, hooksDirMode); mkErr != nil {
		return fmt.Errorf("could not create directory %s: %w", dir, mkErr)
	}

	return afero.WriteFile(l.fs, path, out, checksumFileMode)
}

// writeCursorHookFile reads an existing .cursor/hooks.json (if present), removes
// any lefthook-managed entries (commands starting with "lefthook run "), merges
// in the new entries derived from events, and writes the result back.
//
// Cursor hooks use a flat structure:
//
//	{
//	  "version": 1,
//	  "hooks": {
//	    "<eventName>": [
//	      { "command": "lefthook run <hook>" }
//	    ]
//	  }
//	}
func (l *Lefthook) writeCursorHookFile(path string, events map[string]string) error {
	existing := make(map[string]any)

	data, err := afero.ReadFile(l.fs, path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not read %s: %w", path, err)
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("could not parse %s: %w", path, err)
		}
	}

	mergedHooks := stripCursorLefthookEntries(existing)
	for event, hookName := range events {
		entry := buildCursorHookEntry(hookName)
		if current, ok := mergedHooks[event]; ok {
			if arr, ok := current.([]any); ok {
				mergedHooks[event] = append(arr, entry)
				continue
			}
		}
		mergedHooks[event] = []any{entry}
	}

	existing["version"] = cursorHooksVersion
	existing["hooks"] = mergedHooks

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal %s: %w", path, err)
	}
	out = append(out, '\n')

	dir := filepath.Dir(path)
	if mkErr := l.fs.MkdirAll(dir, hooksDirMode); mkErr != nil {
		return fmt.Errorf("could not create directory %s: %w", dir, mkErr)
	}

	return afero.WriteFile(l.fs, path, out, checksumFileMode)
}

// stripCursorLefthookEntries reads the "hooks" object from an existing Cursor
// hooks map and drops any entries whose command starts with "lefthook run ".
// Non-lefthook entries are preserved as-is.
func stripCursorLefthookEntries(existing map[string]any) map[string]any {
	result := make(map[string]any)

	rawHooks, ok := existing["hooks"]
	if !ok {
		return result
	}

	hooksMap, ok := rawHooks.(map[string]any)
	if !ok {
		return result
	}

	for event, rawEntries := range hooksMap {
		entries, ok := rawEntries.([]any)
		if !ok {
			result[event] = rawEntries
			continue
		}

		var kept []any
		for _, rawEntry := range entries {
			if isCursorLefthookEntry(rawEntry) {
				continue
			}
			kept = append(kept, rawEntry)
		}

		if len(kept) > 0 {
			result[event] = kept
		}
	}

	return result
}

// isCursorLefthookEntry reports whether a Cursor hook entry was generated by lefthook.
func isCursorLefthookEntry(rawEntry any) bool {
	entry, ok := rawEntry.(map[string]any)
	if !ok {
		return false
	}

	cmd, ok := entry["command"].(string)
	return ok && strings.HasPrefix(cmd, lefthookRunPrefix)
}

// buildCursorHookEntry constructs a single Cursor hook that runs `lefthook run <hookName>`.
func buildCursorHookEntry(hookName string) map[string]any {
	return map[string]any{
		"command": lefthookRunPrefix + hookName,
	}
}

// stripLefthookEntries reads the "hooks" object from an existing settings map
// and drops any matcher entries whose sole hook command starts with "lefthook run ".
// Non-lefthook entries are preserved as-is.
func stripLefthookEntries(existing map[string]any) map[string]any {
	result := make(map[string]any)

	rawHooks, ok := existing["hooks"]
	if !ok {
		return result
	}

	hooksMap, ok := rawHooks.(map[string]any)
	if !ok {
		return result
	}

	for event, rawMatchers := range hooksMap {
		matchers, ok := rawMatchers.([]any)
		if !ok {
			result[event] = rawMatchers
			continue
		}

		var kept []any
		for _, rawMatcher := range matchers {
			if isLefthookMatcher(rawMatcher) {
				continue
			}
			kept = append(kept, rawMatcher)
		}

		if len(kept) > 0 {
			result[event] = kept
		}
	}

	return result
}

// isLefthookMatcher reports whether a matcher entry was generated by lefthook.
// A lefthook-generated entry has a "hooks" array where every handler command
// starts with "lefthook run ".
func isLefthookMatcher(rawMatcher any) bool {
	matcher, ok := rawMatcher.(map[string]any)
	if !ok {
		return false
	}

	rawHandlers, ok := matcher["hooks"]
	if !ok {
		return false
	}

	handlers, ok := rawHandlers.([]any)
	if !ok || len(handlers) == 0 {
		return false
	}

	for _, rawHandler := range handlers {
		handler, ok := rawHandler.(map[string]any)
		if !ok {
			return false
		}
		cmd, ok := handler["command"].(string)
		if !ok || !strings.HasPrefix(cmd, lefthookRunPrefix) {
			return false
		}
	}

	return true
}

// buildHookEntry constructs a single matcher group that runs `lefthook run <hookName>`.
func buildHookEntry(hookName string) map[string]any {
	return map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": lefthookRunPrefix + hookName,
			},
		},
	}
}
