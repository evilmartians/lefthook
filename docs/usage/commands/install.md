---
title: "lefthook install"
---

## `lefthook install`

Creates an empty `lefthook.yml` if a configuration file does not exist.

Installs configured hooks to Git hooks.

::: callout info Note
Reinstall is not required when you modify `lefthook.yml`, the configuration file is read every time a git hook is run.
:::

::: callout info Note
NPM package `lefthook` installs the hooks in a postinstall script automatically. For projects not using NPM package run `lefthook install` after cloning the repo.

The postinstall script runs a plain `lefthook install`, so when `core.hooksPath` is set it stops and prints the same message as running the command by hand. Run `lefthook install --force` or `lefthook install --reset-hooks-path` once, deliberately, to resolve it.
:::

### Installing specific hooks

You can install only specific hooks by running `lefthook install <hook-1> <hook-2> ...`.
