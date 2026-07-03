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

	copilotHooksDir     = ".github/hooks"
	copilotHooksFile    = "lefthook.json"
	copilotHooksVersion = 1

	lefthookRunPrefix = "lefthook run "
	lefthookRunSuffix = " run "

	lefthookBinName = "lefthook"
)

var errAIHooksMisconfigured = errors.New("ai hooks misconfigured")

// resolveLefthookBin returns the lefthook executable or command to embed in
// generated AI hook entries. Prefers the config "lefthook" setting, then
// os.Executable(), then the bare "lefthook" name as a last resort.
func resolveLefthookBin(cfg *config.Config) (bin string, quoteBin bool) {
	if cfg != nil {
		if trimmed := strings.TrimSpace(cfg.Lefthook); trimmed != "" {
			return trimmed, false
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return lefthookBinName, false
	}

	return filepath.Clean(exe), true
}

func lefthookRunCommand(bin, hookName string, quoteBin bool) string {
	if quoteBin {
		bin = shellQuotePath(bin)
	}

	return bin + lefthookRunSuffix + hookName
}

func shellQuotePath(path string) string {
	if !strings.Contains(path, " ") {
		return path
	}

	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}

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
	missing = append(missing, checkAIHookReferences("copilot", ai.Copilot, hooks)...)

	if len(missing) > 0 {
		slices.Sort(missing)
		for _, msg := range missing {
			l.logger.Errorf("%s", msg)
		}
		return errAIHooksMisconfigured
	}

	return nil
}

// installAIHooks generates provider-specific settings files for each configured
// LLM provider under the ai: key.
func (l *Lefthook) installAIHooks(ai *config.AI, cfg *config.Config) error {
	bin, quoteBin := resolveLefthookBin(cfg)

	if len(ai.Claude) > 0 {
		path := filepath.Join(l.repo.RootPath, claudeSettingsDir, claudeSettingsFile)
		if err := l.writeAIHookFile(path, ai.Claude, bin, quoteBin); err != nil {
			return fmt.Errorf("could not write Claude settings: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Codex) > 0 {
		path := filepath.Join(l.repo.RootPath, codexHooksDir, codexHooksFile)
		if err := l.writeAIHookFile(path, ai.Codex, bin, quoteBin); err != nil {
			return fmt.Errorf("could not write Codex hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Cursor) > 0 {
		path := filepath.Join(l.repo.RootPath, cursorHooksDir, cursorHooksFile)
		if err := l.writeCursorHookFile(path, ai.Cursor, bin, quoteBin); err != nil {
			return fmt.Errorf("could not write Cursor hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Copilot) > 0 {
		path := filepath.Join(l.repo.RootPath, copilotHooksDir, copilotHooksFile)
		if err := l.writeCopilotHookFile(path, ai.Copilot, bin, quoteBin); err != nil {
			return fmt.Errorf("could not write copilot hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	return nil
}

func (l *Lefthook) writeAIHookFile(path string, events map[string]string, bin string, quoteBin bool) error {
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
		entry := buildHookEntry(hookName, bin, quoteBin)
		if current, ok := mergedHooks[event]; ok {
			if arr, ok := current.([]any); ok {
				mergedHooks[event] = append(arr, entry)
				continue
			}
		}
		mergedHooks[event] = []any{entry}
	}

	existing["hooks"] = mergedHooks

	return l.writeJSONFile(path, existing)
}

func (l *Lefthook) writeCursorHookFile(path string, events map[string]string, bin string, quoteBin bool) error {
	return l.writeFlatHookFile(path, events, cursorHooksVersion, stripCursorLefthookEntries, bin, quoteBin)
}

func (l *Lefthook) writeCopilotHookFile(path string, events map[string]string, bin string, quoteBin bool) error {
	mergedHooks := make(map[string]any, len(events))
	for event, hookName := range events {
		mergedHooks[event] = []any{buildFlatHookEntry(hookName, bin, quoteBin)}
	}

	return l.writeJSONFile(path, map[string]any{
		"version": copilotHooksVersion,
		"hooks":   mergedHooks,
	})
}

func (l *Lefthook) writeJSONFile(path string, data map[string]any) error {
	out, err := json.MarshalIndent(data, "", "  ")
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

func (l *Lefthook) writeFlatHookFile(
	path string,
	events map[string]string,
	version int,
	strip func(map[string]any) map[string]any,
	bin string,
	quoteBin bool,
) error {
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

	mergedHooks := strip(existing)
	for event, hookName := range events {
		entry := buildFlatHookEntry(hookName, bin, quoteBin)
		if current, ok := mergedHooks[event]; ok {
			if arr, ok := current.([]any); ok {
				mergedHooks[event] = append(arr, entry)
				continue
			}
		}
		mergedHooks[event] = []any{entry}
	}

	existing["version"] = version
	existing["hooks"] = mergedHooks

	return l.writeJSONFile(path, existing)
}

func stripCursorLefthookEntries(existing map[string]any) map[string]any {
	return stripFlatLefthookEntries(existing)
}

func stripFlatLefthookEntries(existing map[string]any) map[string]any {
	mergedHooks := make(map[string]any)
	rawHooks, ok := existing["hooks"]
	if !ok {
		return mergedHooks
	}

	hooksMap, ok := rawHooks.(map[string]any)
	if !ok {
		return mergedHooks
	}

	for event, rawEntries := range hooksMap {
		entries, ok := rawEntries.([]any)
		if !ok {
			mergedHooks[event] = rawEntries
			continue
		}

		var kept []any
		for _, rawEntry := range entries {
			if isFlatLefthookEntry(rawEntry) {
				continue
			}
			kept = append(kept, rawEntry)
		}

		if len(kept) > 0 {
			mergedHooks[event] = kept
		}
	}

	return mergedHooks
}

func buildFlatHookEntry(hookName, bin string, quoteBin bool) map[string]any {
	return map[string]any{
		"command": lefthookRunCommand(bin, hookName, quoteBin),
	}
}

func isFlatLefthookEntry(rawEntry any) bool {
	entry, ok := rawEntry.(map[string]any)
	if !ok {
		return false
	}

	cmd, ok := entry["command"].(string)
	return ok && strings.Contains(cmd, lefthookRunPrefix)
}

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
		if !ok || !strings.Contains(cmd, lefthookRunPrefix) {
			return false
		}
	}

	return true
}

// buildHookEntry constructs a single matcher group that runs lefthook for hookName.
func buildHookEntry(hookName, bin string, quoteBin bool) map[string]any {
	return map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": lefthookRunCommand(bin, hookName, quoteBin),
			},
		},
	}
}
