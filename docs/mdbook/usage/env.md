## ENV variables

> ENV variables control some lefthook behavior. Most of them have the alternative CLI or config options.

### `LEFTHOOK`

Use `LEFTHOOK=0 git ...` or `LEFTHOOK=false git ...` to disable lefthook when running git commands.

**Example**

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

When using NPM package `lefthook` in CI, and your CI sets `CI=true` automatically, use `LEFTHOOK=1` or `LEFTHOOK=true` to install hooks in the postinstall script:

**Example**

```bash
LEFTHOOK=1 npm install
LEFTHOOK=1 yarn install
LEFTHOOK=1 pnpm install
```

### `LEFTHOOK_EXCLUDE`

Use `LEFTHOOK_EXCLUDE={list of tags or command names to be excluded}` to skip some commands or scripts by tag or name (for commands only). See the [`exclude_tags`](../configuration/exclude_tags.md) configuration option for more details.

**Example**

```bash
LEFTHOOK_EXCLUDE=ruby,security,lint git commit -am "Skip some tag checks"
```

### `LEFTHOOK_OUTPUT`

Use `LEFTHOOK_OUTPUT={list of output values}` to specify what to print in your output. You can also set `LEFTHOOK_OUTPUT=false` to disable all output except for errors. Refer to the [`output`](../configuration/output.md) configuration option for more details.

**Example**

```bash
$ LEFTHOOK_OUTPUT=summary lefthook run pre-commit
summary: (done in 0.52 seconds)
✔️  lint
```

### `LEFTHOOK_QUIET`

You can skip some outputs printed by lefthook by setting the `LEFTHOOK_QUIET` environment variable. Provide a list of output types to be skipped. See the [`skip_output`](../configuration/skip_output.md) configuration option for more details.

**Example**

```bash
$ LEFTHOOK_QUIET=meta,execution lefthook run pre-commit

  EXECUTE > lint

SUMMARY: (done in 0.01 seconds)
🥊  lint
```

### `LEFTHOOK_VERBOSE`

Set `LEFTHOOK_VERBOSE=1` or `LEFTHOOK_VERBOSE=true` to enable verbose printing.

### `LEFTHOOK_BIN`

Set `LEFTHOOK_BIN` to a location where lefthook is installed to use that instead of trying to detect from the it the PATH or from a package manager.

Useful for cases when:

- lefthook is installed multiple ways, and you want to be explicit about which one is used (example: installed through homebrew, but also is in Gemfile but you are using a ruby version manager like rbenv that prepends it to the path)
- debugging and/or developing lefthook

### `NO_COLOR`

Set `NO_COLOR=true` to disable colored output in lefthook and all subcommands that lefthook calls.

### `CLICOLOR_FORCE`

Set `CLICOLOR_FORCE=true` to force colored output in lefthook and all subcommands.

### `CI`

When using NPM package `lefthook`, set `CI=true` in your CI (if it does not set it automatically) to prevent lefthook from installing hooks in the postinstall script:

```bash
CI=true npm install
CI=true yarn install
CI=true pnpm install
```

> **Note:** Set `LEFTHOOK=1` or `LEFTHOOK=true` to override this behavior and install hooks in the postinstall script (despite `CI=true`).
