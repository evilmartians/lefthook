---
title: "ai"
---

# `ai`

Declare LLM agent hooks directly in `lefthook.yml`. During `lefthook install`, lefthook generates the provider-specific settings file so that the agent calls `lefthook run <hook>` when the event fires.

Each sub-key is a provider name. Its value is a map from the provider's **event name** to a **lefthook hook name** defined elsewhere in the same config.

## Supported providers

| Provider | Generated file | Docs |
|---|---|---|
| `claude` | `.claude/settings.json` | [Claude Code hooks](https://code.claude.com/docs/en/hooks.md) |
| `codex` | `.codex/hooks.json` | [Codex CLI hooks](https://developers.openai.com/codex/hooks) |
| `cursor` | `.cursor/hooks.json` | [Cursor hooks](https://cursor.com/docs/agent/hooks) |
| `copilot` | `.github/hooks/lefthook.json` | [GitHub Copilot hooks](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/use-hooks) |

Keys under each provider must be that provider's hook event names. See the provider's hooks documentation for the supported events and their behaviour.

## Install and uninstall behaviour

Claude, Codex, and Cursor preserve user-authored entries in their settings files. On `lefthook install`, old lefthook-managed entries are replaced with fresh ones derived from the current config. On `lefthook uninstall`, lefthook-managed entries are stripped while user-authored entries stay intact.

Copilot is handled differently: `lefthook install` rewrites `.github/hooks/lefthook.json` from scratch, and `lefthook uninstall` removes that file entirely.

Generated hook commands use the `lefthook` config value when set, otherwise the absolute path of the lefthook binary that ran `install` (via `os.Executable()`), so AI tools do not depend on `lefthook` being on `PATH`.

## Example

```yml
# lefthook.yml

ai:
  claude:
    Stop: validate
    PreToolUse: security-check
  codex:
    Stop: validate
  cursor:
    stop: validate
    preToolUse: security-check
  copilot:
    postToolUse: validate

validate:
  jobs:
    - run: go test ./...

security-check:
  jobs:
    - run: ./scripts/security.sh
```

Running `lefthook install` creates (or updates) `.claude/settings.json`:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "lefthook run validate" }
        ]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [
          { "type": "command", "command": "lefthook run security-check" }
        ]
      }
    ]
  }
}
```

And `.codex/hooks.json`:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "lefthook run validate" }
        ]
      }
    ]
  }
}
```

And `.cursor/hooks.json`:

```json
{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "lefthook run validate" }
    ],
    "preToolUse": [
      { "command": "lefthook run security-check" }
    ]
  }
}
```

And `.github/hooks/lefthook.json`:

```json
{
  "version": 1,
  "hooks": {
    "postToolUse": [
      { "command": "lefthook run validate" }
    ]
  }
}
```
