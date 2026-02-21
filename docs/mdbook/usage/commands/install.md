## `lefthook install`

Creates an empty `lefthook.yml` if a configuration file does not exist.

Installs configured hooks to Git hooks.

> **Note:** Reinstall is not required when you modify `lefthook.yml`, the configuration file is read every time a git hook is run.
> **Note:** NPM package `lefthook` installs the hooks in a postinstall script automatically. For projects not using NPM package run `lefthook install` after cloning the repo.

### Installing specific hooks

You can install only specific hooks by running `lefthook install <hook-1> <hook-2> ...`.
