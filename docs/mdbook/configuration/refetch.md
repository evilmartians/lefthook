## `refetch`

**Default:** `false`

Force remote config refetching on every run. Lefthook will be refetching the specified remote every time it is called.

**Example**

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch: true
```
