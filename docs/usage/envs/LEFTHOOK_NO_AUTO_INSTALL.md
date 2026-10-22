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
