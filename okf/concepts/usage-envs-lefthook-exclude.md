---
type: concept
title: LEFTHOOK_EXCLUDE
source: "https://lefthook.dev/usage/envs/LEFTHOOK_EXCLUDE/"
path: /usage/envs/LEFTHOOK_EXCLUDE/
updated: 2026-08-21
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-21T07:56:19.614Z"
---
---
title: "LEFTHOOK_EXCLUDE"
---

## `LEFTHOOK_EXCLUDE`

Use `LEFTHOOK_EXCLUDE={list of tags or command names to be excluded}` to skip some commands or scripts by tag or name (for commands only). See the [`exclude_tags`](../../configuration/exclude_tags.md) configuration option for more details.

#### Example

```bash
LEFTHOOK_EXCLUDE=ruby,security,lint git commit -am "Skip some tag checks"
```
