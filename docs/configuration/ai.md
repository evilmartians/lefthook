---
title: "ai"
---

# `ai` 🧪 (beta)

> This is a beta feature and still in development.

Declare LLM agent hooks directly in `lefthook.yml`. During `lefthook install`, lefthook generates the provider-specific settings file so that the agent calls `lefthook run <hook>` when the event fires.

Each sub-key is a provider name. Its value is a map from the provider's **event name** to a **lefthook hook name** defined elsewhere in the same config.

## Supported providers

| Provider | Generated file | Docs |
|---|---|---|
| `claude` | `.claude/settings.json` | [Claude Code hooks](https://code.claude.com/docs/en/hooks.md) |
| `codex` | `.codex/hooks.json` | [Codex CLI hooks](https://developers.openai.com/codex/hooks) |
| `cursor` | `.cursor/hooks.json` | [Cursor hooks](https://cursor.com/docs/agent/hooks) |
| `copilot` | `.github/hooks/lefthook.json` | [GitHub Copilot hooks](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/use-hooks) |
| `antigravity` | `.agents/hooks.json` | [Antigravity hooks](https://antigravity.google/docs/hooks) |

Keys under each provider must be that provider's hook event names. See the provider's hooks documentation for the supported events and their behaviour.

## Install and uninstall behaviour

Claude, Codex, and Cursor preserve user-authored entries in their settings files. On `lefthook install`, old lefthook-managed entries are replaced with fresh ones derived from the current config. On `lefthook uninstall`, lefthook-managed entries are stripped while user-authored entries stay intact.

Antigravity preserves user-authored top-level hook groups in `.agents/hooks.json` by managing entries under a dedicated `"lefthook"` top-level key. On `lefthook install`, only the `"lefthook"` key is created or replaced while other top-level hook groups remain untouched. If `ai.antigravity` is empty or removed, `lefthook install` strips the `"lefthook"` key the same way `lefthook uninstall` does. On `lefthook uninstall`, only the `"lefthook"` key is removed; if no other keys remain, `.agents/hooks.json` (and `.agents/` if empty) is deleted.

Antigravity runs hook commands with `.agents/` as the working directory. Plain `lefthook` works from there, but relative paths in the [`lefthook`](./lefthook.md) option resolve from `.agents/`, not from the repository root. Custom fields added inside the managed `"lefthook"` object are overwritten on `lefthook install`; put your own handlers in a separate top-level hook group instead.

Copilot is handled differently: `lefthook install` rewrites `.github/hooks/lefthook.json` from scratch, and `lefthook uninstall` removes that file entirely.

Generated hook commands use the [`lefthook`](./lefthook.md) config value when set. Otherwise they use plain `lefthook`, so it must be on `PATH` for the AI tool. If lefthook is installed as a project dependency, set `lefthook` to e.g. `npx lefthook` or `pnpm exec lefthook`. No machine-specific path is written, so the generated files can be committed and shared.

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
  antigravity:
    Stop: validate
    PreToolUse: security-check

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

And `.agents/hooks.json`:

```json
{
  "lefthook": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "command": "lefthook run security-check",
            "type": "command"
          }
        ]
      }
    ],
    "Stop": [
      {
        "command": "lefthook run validate",
        "type": "command"
      }
    ]
  }
}
```
