---
title: "ref"
---

# `ref`

An optional *branch*, *tag*, or full commit SHA.

A full 40-character commit SHA (or 64-character SHA for SHA-256 repositories) pins the remote to an immutable commit. The remote server must allow fetching commits by SHA.

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

Pin to an immutable commit:

```yml
# lefthook.yml

remotes:
  - git_url: git@github.com:evilmartians/lefthook
    ref: 13d66a090ac6e73f3ff44652761facb02b00e91c
```
