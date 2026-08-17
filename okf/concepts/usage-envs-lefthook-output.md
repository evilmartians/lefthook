---
type: concept
title: LEFTHOOK_OUTPUT
source: "https://lefthook.dev/usage/envs/LEFTHOOK_OUTPUT/"
path: /usage/envs/LEFTHOOK_OUTPUT/
updated: 2026-08-17
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-17T13:11:15.284Z"
---
---
title: "LEFTHOOK_OUTPUT"
---

## `LEFTHOOK_OUTPUT`

Use `LEFTHOOK_OUTPUT={list of output values}` to specify what to print in your output. You can also set `LEFTHOOK_OUTPUT=false` to disable all output except for errors. Refer to the [`output`](../../configuration/output.md) configuration option for more details.

#### Example

```bash
$ LEFTHOOK_OUTPUT=summary lefthook run pre-commit
summary: (done in 0.52 seconds)
✔️  lint
```
