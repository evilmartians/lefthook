package config

// AI holds LLM agent hook integration settings. Each sub-key is a provider name
// whose value is a map from the provider's event name to a lefthook hook name.
// During `lefthook install` the corresponding provider settings file is generated
// (or merged) so that the LLM agent calls `lefthook run <hook>` on that event.
//
// Example lefthook.yml:
//
//	ai:
//	  claude:
//	    Stop: validate
//	    PreToolUse: security-check
//	  codex:
//	    Stop: validate
type AI struct {
	// Claude maps Claude Code event names to lefthook hook names.
	// The generated file is .claude/settings.json.
	Claude map[string]string `json:"claude,omitempty" jsonschema:"description=Claude Code hook mappings (event name -> lefthook hook name). Generates .claude/settings.json." mapstructure:"claude,omitempty" toml:"claude,omitempty" yaml:"claude,omitempty"`

	// Codex maps Codex CLI event names to lefthook hook names.
	// The generated file is .codex/hooks.json.
	Codex map[string]string `json:"codex,omitempty" jsonschema:"description=Codex CLI hook mappings (event name -> lefthook hook name). Generates .codex/hooks.json." mapstructure:"codex,omitempty" toml:"codex,omitempty" yaml:"codex,omitempty"`
}
