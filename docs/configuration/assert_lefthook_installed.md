---
title: "assert_lefthook_installed"
---

# `assert_lefthook_installed`

**Default: `false`**

When set to `true`, fail (with exit status 1) if `lefthook` executable can't be found in $PATH, under node_modules/, as a Ruby gem, or other supported method. This makes sure git hook won't omit `lefthook` rules if `lefthook` ever was installed.

#### Example

```yml
# lefthook.yml

assert_lefthook_installed: true
```

#### Default behavior (`false`)

By default, a missing executable keeps git hooks silent in repositories
without a lefthook config (a shared `core.hooksPath` can serve many
repositories). When the repository does contain a lefthook config file
(`lefthook.yml`, `lefthook.toml`, a `-local` variant, or a path given via
`LEFTHOOK_CONFIG`), a missing executable aborts the hook with exit status 1
instead of silently skipping the configured rules.

The `--no-verify` git argument and the `LEFTHOOK=0` environment variable
still skip the hook in both modes.
