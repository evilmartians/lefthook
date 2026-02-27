---
title: "refetch"
---

# `refetch`

**Default:** `false`

Force remote config refetching on every run. Lefthook will be refetching the specified remote every time it is called.

See [`refetch_frequency`](./refetch_frequency.md) for more flexible refetching options and additional considerations.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch: true
```
