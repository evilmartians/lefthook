---
title: "stash_unstaged_changes"
---

# `stash_unstaged_changes`

**Default: `true` for `pre-commit`**

Controls whether lefthook hides unstaged hunks from partially staged files before running the `pre-commit` hook.

When set to `false`, lefthook keeps the current on-disk contents of partially staged files visible during `pre-commit` execution. File selection does not change: `pre-commit` still uses the staged file set, and explicit templates like [`{staged_files}`](./run.md#staged_files) keep their usual meaning.

::: callout info Note
Works only for the `pre-commit` hook.
:::

::: callout info Note
When this is `false`, [`stage_fixed`](./stage_fixed.md) is disabled for `pre-commit` jobs.
:::

If you need to customize which files are passed to a job, use the hook-level [`files`](./files-global.md) option or job-level [`files`](./files.md) option.

#### Example

```yml
# lefthook.yml

pre-commit:
  stash_unstaged_changes: false
  commands:
    lint:
      run: yarn eslint {staged_files}
```