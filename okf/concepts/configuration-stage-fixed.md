---
type: concept
title: stage_fixed
source: "https://lefthook.dev/configuration/stage_fixed/"
path: /configuration/stage_fixed/
updated: 2026-08-28
okf:
  generated_by: "@docmd/plugin-okf"
  generated_at: "2026-08-28T10:14:00.433Z"
---
---
title: "stage_fixed"
---

# `stage_fixed`

**Default: `false`**

::: callout info Note
Works **only** for the `pre-commit` hook.
:::

When set to `true` lefthook will automatically call `git add` on files after running the command or script. For a command if [`files`](./files.md) option was specified, the specified command will be used to retrieve files for `git add`. For scripts and commands without [`files`](./files.md) option `{staged_files}` template will be used. All filters ([`glob`](./glob.md), [`exclude`](./exclude.md)) will be applied if specified.

If the `git add` call fails, the hook fails too. Otherwise the commit would silently go through with the unfixed content.

#### Example

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      run: npm run lint --fix {staged_files}
      stage_fixed: true
```
