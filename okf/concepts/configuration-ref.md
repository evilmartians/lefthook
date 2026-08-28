---
type: concept
title: ref
source: "https://lefthook.dev/configuration/ref/"
path: /configuration/ref/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:14:00.427Z"
---
---
title: "ref"
---

# `ref`

An optional *branch* or *tag* name.

::: callout info Note
If you initially had `ref` option, ran `lefthook install`, and then removed it, lefthook won't decide which branch/tag to use as a ref. So, if you added it once, please, use it always to avoid issues in local setups.
:::

See also [`refetch_frequency`](./refetch_frequency.md).

#### Example

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
```
