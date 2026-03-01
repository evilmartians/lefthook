---
title: "refetch_frequency"
---

# `refetch_frequency`

**Default:** Not set

Specifies how frequently Lefthook should refetch the remote configuration. This can be set to `always`, `never` or a time duration like `24h`, `30m`, etc.

- When set to `always`, Lefthook will always refetch the remote configuration on each run.
- When set to a duration (e.g., `24h`), Lefthook will check the last fetch time and refetch the configuration only if the specified amount of time has passed.
- When set to `never` or not set, Lefthook will not fetch from remote.

It is recommended to configure remotes that point to mutable references
(including ones without a `ref`) to be refetched with some frequency appropriate for the project.

Failure to fetch does not cause an error, but just a warning message.
If a successfully fetched previous configuration exists, it will be used.
Otherwise, the remote will be ignored.

#### Example

```yml
# lefthook.yml

remotes:
  - git_url: https://github.com/evilmartians/lefthook
    refetch_frequency: 24h # Refetches once every 24 hours
```

> WARNING
> If `refetch` is set to `true`, it overrides any setting in `refetch_frequency`.
