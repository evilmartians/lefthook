---
type: concept
title: LEFTHOOK_NO_AUTO_INSTALL
source: "https://lefthook.dev/usage/envs/LEFTHOOK_NO_AUTO_INSTALL/"
path: /usage/envs/LEFTHOOK_NO_AUTO_INSTALL/
updated: 2026-11-12
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-11-12T12:17:15.705Z"
---
---
title: "LEFTHOOK_NO_AUTO_INSTALL"
---

## `LEFTHOOK_NO_AUTO_INSTALL`

Use `LEFTHOOK_NO_AUTO_INSTALL=1 git ...` or `LEFTHOOK_NO_AUTO_INSTALL=true git ...` to keep lefthook from updating git hooks when running git commands. Unlike the [`no_auto_install`](../../configuration/no_auto_install.md) configuration option, it applies only to the commands that have it set.

Useful for tools that run git on someone else's behalf (wrappers, dev containers, CI helpers) and must not change installed hooks.

#### Example

```bash
LEFTHOOK_NO_AUTO_INSTALL=1 git commit -m "Hooks not synchronized"
```
