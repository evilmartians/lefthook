package config

// AI holds LLM agent hook integration settings. Each sub-key is a provider name
// whose value is a map from the provider's event name to a lefthook hook name.
// During `lefthook install` the corresponding provider settings file is generated
// so that the LLM agent calls `lefthook run <hook>` on that event.
//
// Example lefthook.yml:
//
//	ai:
//	  claude:
//	    Stop: validate
//	    PreToolUse: security-check
//	  codex:
//	    Stop: validate
//	  cursor:
//	    stop: validate
//	  copilot:
//	    postToolUse: validate
type AI struct {
	// Claude maps Claude Code event names to lefthook hook names.
	// The generated file is .claude/settings.json.
	Claude map[string]string `json:"claude,omitempty" jsonschema:"description=Claude Code hook mappings (event name -> lefthook hook name). Generates .claude/settings.json." mapstructure:"claude,omitempty" toml:"claude,omitempty" yaml:"claude,omitempty"`

	// Codex maps Codex CLI event names to lefthook hook names.
	// The generated file is .codex/hooks.json.
	Codex map[string]string `json:"codex,omitempty" jsonschema:"description=Codex CLI hook mappings (event name -> lefthook hook name). Generates .codex/hooks.json." mapstructure:"codex,omitempty" toml:"codex,omitempty" yaml:"codex,omitempty"`

	// Cursor maps Cursor hook event names to lefthook hook names.
	// The generated file is .cursor/hooks.json.
	Cursor map[string]string `json:"cursor,omitempty" jsonschema:"description=Cursor hook mappings (event name -> lefthook hook name). Generates .cursor/hooks.json." mapstructure:"cursor,omitempty" toml:"cursor,omitempty" yaml:"cursor,omitempty"`

	// Copilot is for GitHub Copilot agent.
	// The generated file is .github/hooks/lefthook.json.
	Copilot map[string]string `json:"copilot,omitempty" jsonschema:"description=Copilot hooks mapping. Generates .github/hooks/lefthook.json." mapstructure:"copilot,omitempty" toml:"copilot,omitempty" yaml:"copilot,omitempty"`
}
