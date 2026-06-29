---
title: "ai"
---

# `ai`

Declare LLM agent hooks directly in `lefthook.yml`. During `lefthook install`, lefthook generates (or merges into) the provider-specific settings file so that the agent calls `lefthook run <hook>` when the event fires.

Each sub-key is a provider name. Its value is a map from the provider's **event name** to a **lefthook hook name** defined elsewhere in the same config.

## Supported providers

| Provider | Generated file | Docs |
|---|---|---|
| `claude` | `.claude/settings.json` | [Claude Code hooks](https://code.claude.com/docs/en/hooks.md) |
| `codex` | `.codex/hooks.json` | [Codex CLI hooks](https://developers.openai.com/codex/hooks) |
| `cursor` | `.cursor/hooks.json` | [Cursor hooks](https://cursor.com/docs/agent/hooks) |

Keys under each provider must be that provider's hook event names. See the provider's hooks documentation for the supported events and their behaviour.

## Merge behaviour

Lefthook reads any existing provider file and **preserves** entries that were not written by lefthook. On every `lefthook install` run, stale lefthook-generated entries are replaced with fresh ones derived from the current config, so the file stays up to date without accumulating duplicates.

Running `lefthook uninstall` reverses this: lefthook-generated entries are stripped from the provider files while user-authored entries are preserved. A provider file that contained only lefthook-generated entries is removed entirely.

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
