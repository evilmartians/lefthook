---
title: "ref"
---

# `ref`

An optional *branch* or *tag* name.

> **Note:** If you initially had `ref` option, ran `lefthook install`, and then removed it, lefthook won't decide which branch/tag to use as a ref. So, if you added it once, please, use it always to avoid issues in local setups.

See also [`refetch_frequency`](./refetch_frequency.md).

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: v1.0.0
```
