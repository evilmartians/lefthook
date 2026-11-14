package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
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

	antigravityHooksDir  = ".agents"
	antigravityHooksFile = "hooks.json"
	antigravityHooksKey  = "lefthook"
	antigravityMatcher   = "*"

	lefthookRunPrefix = "lefthook run "
	lefthookRunSuffix = " run "

	lefthookBinName = "lefthook"
)

type antigravityEventFormat uint8

const (
	antigravityEventFlat antigravityEventFormat = iota
	antigravityEventMatcher
)

var (
	errAIHooksMisconfigured = errors.New("ai hooks misconfigured")

	// antigravityEvents maps supported Antigravity hook event names to the JSON
	// entry format required in .agents/hooks.json.
	antigravityEvents = map[string]antigravityEventFormat{
		"SessionStart":   antigravityEventFlat,
		"PreToolUse":     antigravityEventMatcher,
		"PostToolUse":    antigravityEventMatcher,
		"PreInvocation":  antigravityEventFlat,
		"PostInvocation": antigravityEventFlat,
		"Stop":           antigravityEventFlat,
	}
)

// resolveLefthookBin returns the command to embed in generated AI hook entries.
// These settings files are usually committed, so the command must not depend on
// where lefthook is installed on the current machine. Prefers the config
// "lefthook" setting, then the bare "lefthook" name.
func resolveLefthookBin(cfg *config.Config) string {
	if cfg != nil {
		if trimmed := strings.TrimSpace(cfg.Lefthook); trimmed != "" {
			return trimmed
		}
	}

	return lefthookBinName
}

func lefthookRunCommand(bin, hookName string) string {
	return bin + lefthookRunSuffix + hookName
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

func checkAntigravityEvents(events map[string]string) []string {
	var invalid []string
	for event := range events {
		if _, ok := antigravityEvents[event]; !ok {
			invalid = append(invalid, fmt.Sprintf("ai.antigravity.%s: unsupported event", event))
		}
	}
	return invalid
}

// validateAIHooks ensures every hook name referenced under the ai: key points to
// a hook that is actually defined in the config and validates provider event names.
// This catches typos early instead of silently writing a `lefthook run <typo>`
// command that fails when the agent event fires.
func (l *Lefthook) validateAIHooks(ai *config.AI, hooks map[string]*config.Hook) error {
	issues := checkAIHookReferences("claude", ai.Claude, hooks)
	issues = append(issues, checkAIHookReferences("codex", ai.Codex, hooks)...)
	issues = append(issues, checkAIHookReferences("cursor", ai.Cursor, hooks)...)
	issues = append(issues, checkAIHookReferences("copilot", ai.Copilot, hooks)...)
	issues = append(issues, checkAIHookReferences("antigravity", ai.Antigravity, hooks)...)
	issues = append(issues, checkAntigravityEvents(ai.Antigravity)...)

	if len(issues) > 0 {
		slices.Sort(issues)
		for _, issue := range issues {
			l.logger.Errorf("%s", issue)
		}
		return errAIHooksMisconfigured
	}

	return nil
}

// installAIHooks generates provider-specific settings files for each configured
// LLM provider under the ai: key.
func (l *Lefthook) installAIHooks(ai *config.AI, cfg *config.Config) error {
	bin := resolveLefthookBin(cfg)

	if len(ai.Claude) > 0 {
		path := filepath.Join(l.repo.Paths.Root, claudeSettingsDir, claudeSettingsFile)
		if err := l.writeAIHookFile(path, ai.Claude, bin); err != nil {
			return fmt.Errorf("could not write Claude settings: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Codex) > 0 {
		path := filepath.Join(l.repo.Paths.Root, codexHooksDir, codexHooksFile)
		if err := l.writeAIHookFile(path, ai.Codex, bin); err != nil {
			return fmt.Errorf("could not write Codex hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Cursor) > 0 {
		path := filepath.Join(l.repo.Paths.Root, cursorHooksDir, cursorHooksFile)
		if err := l.writeCursorHookFile(path, ai.Cursor, bin); err != nil {
			return fmt.Errorf("could not write Cursor hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Copilot) > 0 {
		path := filepath.Join(l.repo.Paths.Root, copilotHooksDir, copilotHooksFile)
		if err := l.writeCopilotHookFile(path, ai.Copilot, bin); err != nil {
			return fmt.Errorf("could not write copilot hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	if len(ai.Antigravity) > 0 {
		path := filepath.Join(l.repo.Paths.Root, antigravityHooksDir, antigravityHooksFile)
		if err := l.writeAntigravityHookFile(path, ai.Antigravity, bin); err != nil {
			return fmt.Errorf("could not write Antigravity hooks: %w", err)
		}
		l.logger.Infof("Updated %s", path)
	}

	return nil
}

func (l *Lefthook) writeAIHookFile(path string, events map[string]string, bin string) error {
	existing, _, err := l.readJSONFile(path)
	if err != nil {
		return err
	}

	mergedHooks := stripLefthookEntries(existing)
	for event, hookName := range events {
		entry := buildHookEntry(hookName, bin)
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

func (l *Lefthook) writeCursorHookFile(path string, events map[string]string, bin string) error {
	return l.writeFlatHookFile(path, events, cursorHooksVersion, stripCursorLefthookEntries, bin)
}

func (l *Lefthook) writeCopilotHookFile(path string, events map[string]string, bin string) error {
	mergedHooks := make(map[string]any, len(events))
	for event, hookName := range events {
		mergedHooks[event] = []any{buildFlatHookEntry(hookName, bin)}
	}

	return l.writeJSONFile(path, map[string]any{
		"version": copilotHooksVersion,
		"hooks":   mergedHooks,
	})
}

func (l *Lefthook) writeAntigravityHookFile(path string, events map[string]string, bin string) error {
	lefthookHooks := make(map[string]any, len(events))
	for _, event := range slices.Sorted(maps.Keys(events)) {
		format, ok := antigravityEvents[event]
		if !ok {
			return fmt.Errorf("%w: unsupported antigravity event %q", errAIHooksMisconfigured, event)
		}
		lefthookHooks[event] = buildAntigravityEventEntries(format, events[event], bin)
	}

	existing, _, err := l.readJSONFile(path)
	if err != nil {
		return err
	}

	existing[antigravityHooksKey] = lefthookHooks

	return l.writeJSONFile(path, existing)
}

// readJSONFile reads and unmarshals a JSON object file. The bool return reports
// whether the file existed and was non-empty.
func (l *Lefthook) readJSONFile(path string) (map[string]any, bool, error) {
	data, err := afero.ReadFile(l.fs, path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]any), false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("could not read %s: %w", path, err)
	}
	if len(data) == 0 {
		return make(map[string]any), false, nil
	}

	existing := make(map[string]any)
	if err = json.Unmarshal(data, &existing); err != nil {
		return nil, false, fmt.Errorf("could not parse %s: %w", path, err)
	}
	if existing == nil {
		existing = make(map[string]any)
	}

	return existing, true, nil
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
) error {
	existing, _, err := l.readJSONFile(path)
	if err != nil {
		return err
	}

	mergedHooks := strip(existing)
	for event, hookName := range events {
		entry := buildFlatHookEntry(hookName, bin)
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

func buildFlatHookEntry(hookName, bin string) map[string]any {
	return map[string]any{
		"command": lefthookRunCommand(bin, hookName),
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

func buildCommandHandler(hookName, bin string) map[string]any {
	return map[string]any{
		"type":    "command",
		"command": lefthookRunCommand(bin, hookName),
	}
}

func buildAntigravityEventEntries(format antigravityEventFormat, hookName, bin string) []any {
	handler := buildCommandHandler(hookName, bin)
	if format == antigravityEventMatcher {
		return []any{
			map[string]any{
				"matcher": antigravityMatcher,
				"hooks":   []any{handler},
			},
		}
	}
	return []any{handler}
}

// buildHookEntry constructs a single matcher group that runs lefthook for hookName.
func buildHookEntry(hookName, bin string) map[string]any {
	return map[string]any{
		"hooks": []any{
			buildCommandHandler(hookName, bin),
		},
	}
}
