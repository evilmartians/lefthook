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
